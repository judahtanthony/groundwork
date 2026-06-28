// Package decision implements ticket-attached durable decision records
// (docs/contracts/decision-records.md, ADR 0051). These are the durable
// operational memory that explains why a ticket is blocked, in review, or ready
// for rework, and they let pending input/approval/decision queues rebuild after
// .groundwork/state.sqlite is purged (ADR 0053).
//
// The on-disk form is a newline-delimited JSON sidecar
// (.groundwork/tickets/<id>/decisions.ndjson). Encoding is canonical (ADR 0020):
// compact JSON in fixed key order, UTC timestamps, and records emitted in a
// deterministic order so a rebuilt store re-exports byte-for-byte.
package decision

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Event types (docs/contracts/decision-records.md).
const (
	EventDecisionRequested = "decision_requested"
	EventDecisionResponded = "decision_responded"
	EventInputRequested    = "input_requested"
	EventInputAnswered     = "input_answered"
	EventApprovalRequested = "approval_requested"
	EventApprovalDecided   = "approval_decided"
	EventReworkRequested   = "rework_requested"
	EventRecoveryNeeded    = "recovery_needed"
)

// Statuses (docs/contracts/decision-records.md). The current status for a
// durable request id is the latest non-superseded event for that id.
const (
	StatusPending    = "pending"
	StatusAnswered   = "answered"
	StatusAccepted   = "accepted"
	StatusRejected   = "rejected"
	StatusSuperseded = "superseded"
	StatusRecovered  = "recovered"
)

// PolicyInputs records the gate inputs that produced a request, so a rebuilt
// queue can re-derive the gate decision without the live run.
type PolicyInputs struct {
	Action     string `json:"action,omitempty"`
	RiskClass  string `json:"risk_class,omitempty"`
	Reversible *bool  `json:"reversible,omitempty"`
}

// Option is a single selectable response on a decision/approval request.
type Option struct {
	ID    string `json:"id"`
	Label string `json:"label"`
}

// Record is one append-only semantic event in a ticket's decisions.ndjson.
// Field order here is the canonical key order for export (ADR 0020). Optional
// fields use omitempty so absent data is omitted identically on every write.
// The durable id is stable across rebuilds; runtime approval/queue handles are
// not stored here.
type Record struct {
	ID                 string        `json:"id,omitempty"`
	Sequence           int           `json:"sequence"`
	EventType          string        `json:"event_type"`
	TicketID           string        `json:"ticket_id"`
	RunID              string        `json:"run_id,omitempty"`
	RequestType        string        `json:"request_type,omitempty"`
	WorkType           string        `json:"work_type,omitempty"`
	Status             string        `json:"status"`
	RequestedBy        string        `json:"requested_by,omitempty"`
	RequestedAt        string        `json:"requested_at,omitempty"`
	RequestedActor     string        `json:"requested_actor,omitempty"`
	RequiredRoles      []string      `json:"required_roles,omitempty"`
	PolicyInputs       *PolicyInputs `json:"policy_inputs,omitempty"`
	Statement          string        `json:"statement,omitempty"`
	HandoffSummary     string        `json:"handoff_summary,omitempty"`
	Options            []Option      `json:"options,omitempty"`
	Recommendation     string        `json:"recommendation,omitempty"`
	RiskNotes          string        `json:"risk_notes,omitempty"`
	ScopeNotes         string        `json:"scope_notes,omitempty"`
	RelatedFiles       []string      `json:"related_files,omitempty"`
	Artifacts          []string      `json:"artifacts,omitempty"`
	Checkpoints        []string      `json:"checkpoints,omitempty"`
	DependsOn          []string      `json:"depends_on,omitempty"`
	Response           string        `json:"response,omitempty"`
	DecidedBy          string        `json:"decided_by,omitempty"`
	DecidedAt          string        `json:"decided_at,omitempty"`
	FollowUp           []string      `json:"follow_up,omitempty"`
	CanonUpdatesNeeded []string      `json:"canon_updates_needed,omitempty"`
}

// Validate checks the contract's required fields: event_type, ticket_id,
// status, and a human-readable statement or handoff_summary.
func (r *Record) Validate() error {
	if r.EventType == "" {
		return fmt.Errorf("decision record: event_type is required")
	}
	if r.TicketID == "" {
		return fmt.Errorf("decision record %q: ticket_id is required", r.EventType)
	}
	if r.Status == "" {
		return fmt.Errorf("decision record %q: status is required", r.EventType)
	}
	if strings.TrimSpace(r.Statement) == "" && strings.TrimSpace(r.HandoffSummary) == "" {
		return fmt.Errorf("decision record %q: statement or handoff_summary is required", r.EventType)
	}
	return nil
}

// SidecarName is the ticket-relative sidecar filename.
const SidecarName = "decisions.ndjson"

// Path returns the decisions sidecar path for a ticket under ticketsDir.
func Path(ticketsDir, ticketID string) string {
	return filepath.Join(ticketsDir, ticketID, SidecarName)
}

// sortRecords orders records deterministically for byte-stable export: by
// Sequence (the authoritative append order), then id, then event type as
// tiebreakers. Sorting in place.
func sortRecords(records []Record) {
	sort.SliceStable(records, func(i, j int) bool {
		a, b := records[i], records[j]
		if a.Sequence != b.Sequence {
			return a.Sequence < b.Sequence
		}
		if a.ID != b.ID {
			return a.ID < b.ID
		}
		return a.EventType < b.EventType
	})
}

// Encode renders records as canonical NDJSON: one compact JSON object per line
// in deterministic order, with a trailing newline after each record. Encoding
// is idempotent — Encode(Decode(Encode(x))) == Encode(x).
func Encode(records []Record) ([]byte, error) {
	ordered := append([]Record(nil), records...)
	sortRecords(ordered)
	var b bytes.Buffer
	for i := range ordered {
		if err := ordered[i].Validate(); err != nil {
			return nil, err
		}
		line, err := json.Marshal(&ordered[i])
		if err != nil {
			return nil, err
		}
		b.Write(line)
		b.WriteByte('\n')
	}
	return b.Bytes(), nil
}

// Decode parses canonical NDJSON back into records. Blank lines are tolerated;
// every non-blank line must be a valid record object.
func Decode(data []byte) ([]Record, error) {
	var records []Record
	sc := bufio.NewScanner(bytes.NewReader(data))
	sc.Buffer(make([]byte, 0, 64*1024), 4*1024*1024)
	line := 0
	for sc.Scan() {
		line++
		raw := bytes.TrimSpace(sc.Bytes())
		if len(raw) == 0 {
			continue
		}
		var r Record
		dec := json.NewDecoder(bytes.NewReader(raw))
		dec.DisallowUnknownFields()
		if err := dec.Decode(&r); err != nil {
			return nil, fmt.Errorf("decisions.ndjson line %d: %w", line, err)
		}
		if err := r.Validate(); err != nil {
			return nil, fmt.Errorf("decisions.ndjson line %d: %w", line, err)
		}
		records = append(records, r)
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	return records, nil
}

// Read loads a ticket's decisions sidecar. The bool is false when no sidecar
// exists (the file is optional for tickets with no durable history).
func Read(ticketsDir, ticketID string) ([]Record, bool, error) {
	data, err := os.ReadFile(Path(ticketsDir, ticketID))
	if os.IsNotExist(err) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	recs, err := Decode(data)
	if err != nil {
		return nil, false, err
	}
	return recs, true, nil
}

// Write persists a ticket's decisions sidecar (the authoritative copy). With no
// records the sidecar is removed so an empty file is never committed; the
// sidecar is optional per the contract. The ticket directory is created if
// needed.
func Write(ticketsDir, ticketID string, records []Record) error {
	path := Path(ticketsDir, ticketID)
	if len(records) == 0 {
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			return err
		}
		return nil
	}
	data, err := Encode(records)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}
