// Package client is the HTTP client for the gw coordinator API. The CLI uses it
// to route mutating commands through a running coordinator (ADR 0031). It maps
// the server's JSON error codes back to the store's sentinel errors so callers
// can branch on errors.Is(err, sqlite.ErrNotFound) regardless of transport.
package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"groundwork/internal/store/sqlite"
	"groundwork/internal/ticket"
)

// Client talks to a coordinator at a host:port address.
type Client struct {
	base string
	http *http.Client
}

// New returns a client for the coordinator at addr (host:port).
func New(addr string) *Client {
	return &Client{
		base: "http://" + addr,
		http: &http.Client{Timeout: 5 * time.Second},
	}
}

// Healthy reports whether the coordinator answers its health endpoint. A
// connection refusal (no server running) returns false quickly.
func (c *Client) Healthy() bool {
	resp, err := c.http.Get(c.base + "/healthz")
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

// GetTicket fetches one ticket.
func (c *Client) GetTicket(id string) (*ticket.Ticket, error) {
	var t ticket.Ticket
	if err := c.do(http.MethodGet, "/api/v1/tickets/"+id, nil, &t); err != nil {
		return nil, err
	}
	return &t, nil
}

// CreateTicket creates a ticket; the assigned id and timestamps are written back
// into t. The actor argument is ignored over HTTP: the coordinator records the
// owner actor for API-initiated mutations (single-user v1).
func (c *Client) CreateTicket(t *ticket.Ticket, actor string) error {
	return c.do(http.MethodPost, "/api/v1/tickets", t, t)
}

// UpdateTicket persists the mutable fields of t (the caller performs the
// read-modify-write; PATCH replaces the mutable resource representation).
func (c *Client) UpdateTicket(t *ticket.Ticket, actor string) error {
	return c.do(http.MethodPatch, "/api/v1/tickets/"+t.ID, t, t)
}

// TransitionTicket changes a ticket's status.
func (c *Client) TransitionTicket(id string, to ticket.Status, actor string) error {
	return c.do(http.MethodPost, "/api/v1/tickets/"+id+"/transition",
		map[string]string{"status": string(to)}, nil)
}

// AddDependency records that fromID depends on toID.
func (c *Client) AddDependency(fromID, toID, actor string) error {
	return c.do(http.MethodPost, "/api/v1/tickets/"+fromID+"/dependencies",
		map[string]string{"depends_on": toID}, nil)
}

// RemoveDependency deletes the edge fromID -> toID.
func (c *Client) RemoveDependency(fromID, toID, actor string) error {
	return c.do(http.MethodDelete, "/api/v1/tickets/"+fromID+"/dependencies/"+toID, nil, nil)
}

// Reparent moves id under newParentID.
func (c *Client) Reparent(id, newParentID, actor string) error {
	return c.do(http.MethodPost, "/api/v1/tickets/"+id+"/reparent",
		map[string]string{"parent": newParentID}, nil)
}

// ListRuns returns runs newest-first.
func (c *Client) ListRuns() ([]*sqlite.Run, error) {
	var runs []*sqlite.Run
	if err := c.do(http.MethodGet, "/api/v1/runs", nil, &runs); err != nil {
		return nil, err
	}
	return runs, nil
}

// GetRun returns one run.
func (c *Client) GetRun(id string) (*sqlite.Run, error) {
	var run sqlite.Run
	if err := c.do(http.MethodGet, "/api/v1/runs/"+id, nil, &run); err != nil {
		return nil, err
	}
	return &run, nil
}

// RunEvents returns a run's event log.
func (c *Client) RunEvents(id string) ([]sqlite.RunEvent, error) {
	var events []sqlite.RunEvent
	if err := c.do(http.MethodGet, "/api/v1/runs/"+id+"/events", nil, &events); err != nil {
		return nil, err
	}
	return events, nil
}

// RunOnce triggers a scheduling attempt for one node (gw run once).
func (c *Client) RunOnce(ticketID string) (*sqlite.Run, error) {
	var run sqlite.Run
	if err := c.do(http.MethodPost, "/api/v1/runs", map[string]string{"ticket_id": ticketID}, &run); err != nil {
		return nil, err
	}
	return &run, nil
}

// RunNext triggers one scheduler tick and returns how many runs started.
func (c *Client) RunNext() (int, error) {
	var out struct {
		Started int `json:"started"`
	}
	if err := c.do(http.MethodPost, "/api/v1/runs", map[string]string{}, &out); err != nil {
		return 0, err
	}
	return out.Started, nil
}

// PauseRun, ResumeRun, and CancelRun are the live run-control transitions.
func (c *Client) PauseRun(id string) (*sqlite.Run, error)  { return c.runControl(id, "pause") }
func (c *Client) ResumeRun(id string) (*sqlite.Run, error) { return c.runControl(id, "resume") }
func (c *Client) CancelRun(id string) (*sqlite.Run, error) { return c.runControl(id, "cancel") }

func (c *Client) runControl(id, op string) (*sqlite.Run, error) {
	var run sqlite.Run
	if err := c.do(http.MethodPost, "/api/v1/runs/"+id+"/"+op, nil, &run); err != nil {
		return nil, err
	}
	return &run, nil
}

// ListApprovals returns approvals, optionally filtered by status.
func (c *Client) ListApprovals(status string) ([]*sqlite.Approval, error) {
	path := "/api/v1/approvals"
	if status != "" {
		path += "?status=" + status
	}
	var out []*sqlite.Approval
	if err := c.do(http.MethodGet, path, nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetApproval returns one approval.
func (c *Client) GetApproval(id string) (*sqlite.Approval, error) {
	var a sqlite.Approval
	if err := c.do(http.MethodGet, "/api/v1/approvals/"+id, nil, &a); err != nil {
		return nil, err
	}
	return &a, nil
}

// DecideApproval records a decision (op = approve|reject|clarify).
func (c *Client) DecideApproval(id, op, reason string) (*sqlite.Approval, error) {
	var a sqlite.Approval
	if err := c.do(http.MethodPost, "/api/v1/approvals/"+id+"/"+op, map[string]string{"reason": reason}, &a); err != nil {
		return nil, err
	}
	return &a, nil
}

// DecomposeTicket records a decomposition proposal, returning the pending
// approval and the created (backlog) child ids.
func (c *Client) DecomposeTicket(id, contract string, children []sqlite.ChildSpec) (*sqlite.Approval, []string, error) {
	body := map[string]any{"children": children}
	if contract != "" {
		body["contract"] = json.RawMessage(contract)
	}
	var out struct {
		Approval *sqlite.Approval `json:"approval"`
		ChildIDs []string         `json:"child_ids"`
	}
	if err := c.do(http.MethodPost, "/api/v1/tickets/"+id+"/decompose", body, &out); err != nil {
		return nil, nil, err
	}
	return out.Approval, out.ChildIDs, nil
}

// ListValidations returns a node's validation results.
func (c *Client) ListValidations(ticketID string) ([]*sqlite.ValidationResult, error) {
	var out []*sqlite.ValidationResult
	if err := c.do(http.MethodGet, "/api/v1/tickets/"+ticketID+"/validations", nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// RecordValidation records a validation result through the coordinator.
func (c *Client) RecordValidation(ticketID string, v sqlite.ValidationResult) (*sqlite.ValidationResult, error) {
	body := map[string]string{"name": v.Name, "command": v.Command, "status": v.Status, "artifact_path": v.ArtifactPath}
	var out sqlite.ValidationResult
	if err := c.do(http.MethodPost, "/api/v1/tickets/"+ticketID+"/validations", body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// LandResult is the outcome of a land request: either landed, or a pending
// human approval was opened.
type LandResult struct {
	Landed   bool             `json:"landed"`
	Ticket   *ticket.Ticket   `json:"ticket,omitempty"`
	Approval *sqlite.Approval `json:"approval,omitempty"`
}

// LandTicket requests landing through the gate: it lands immediately when policy
// auto-approves (or override is set), otherwise it opens a pending human
// approval and returns it.
func (c *Client) LandTicket(ticketID string, override bool) (*LandResult, error) {
	var out LandResult
	if err := c.do(http.MethodPost, "/api/v1/tickets/"+ticketID+"/land", map[string]bool{"override": override}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// EscalateTicket records an escalation and opens a re-plan approval.
func (c *Client) EscalateTicket(id, reason string) (*sqlite.Approval, error) {
	var a sqlite.Approval
	if err := c.do(http.MethodPost, "/api/v1/tickets/"+id+"/escalate", map[string]string{"reason": reason}, &a); err != nil {
		return nil, err
	}
	return &a, nil
}

// do performs a request, encoding body (if non-nil) and decoding a 2xx response
// into out (if non-nil). Non-2xx responses are mapped to sentinel errors.
func (c *Client) do(method, path string, body, out any) error {
	var reader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return err
		}
		reader = bytes.NewReader(b)
	}
	req, err := http.NewRequest(method, c.base+path, reader)
	if err != nil {
		return err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return decodeError(resp)
	}
	if out != nil {
		return json.NewDecoder(resp.Body).Decode(out)
	}
	return nil
}

// errorEnvelope mirrors docs/contracts/http-api.md.
type errorEnvelope struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

// APIError carries a coordinator error code/message for cases without a
// matching store sentinel.
type APIError struct {
	Code    string
	Message string
}

func (e *APIError) Error() string { return e.Message }

// decodeError reads the error envelope and maps known codes to the store
// sentinels so CLI error switches behave the same over HTTP as over the store.
func decodeError(resp *http.Response) error {
	var env errorEnvelope
	if err := json.NewDecoder(resp.Body).Decode(&env); err != nil {
		return fmt.Errorf("coordinator returned %s", resp.Status)
	}
	switch env.Error.Code {
	case "not_found":
		return sqlite.ErrNotFound
	case "illegal_transition":
		return fmt.Errorf("%w: %s", sqlite.ErrIllegalTransition, env.Error.Message)
	case "self_dependency":
		return sqlite.ErrSelfDependency
	case "dependency_cycle":
		return fmt.Errorf("%w: %s", sqlite.ErrDependencyCycle, env.Error.Message)
	case "self_parent":
		return sqlite.ErrSelfParent
	case "parent_cycle":
		return fmt.Errorf("%w: %s", sqlite.ErrParentCycle, env.Error.Message)
	case "empty_title":
		return sqlite.ErrEmptyTitle
	default:
		return &APIError{Code: env.Error.Code, Message: env.Error.Message}
	}
}
