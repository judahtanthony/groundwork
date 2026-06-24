package server

import (
	"encoding/json"
	"fmt"
	"net/http"

	"groundwork/internal/actor"
	"groundwork/internal/approval"
	"groundwork/internal/policy"
	"groundwork/internal/store/sqlite"
)

// handleApprovalList returns approvals, optionally filtered by ?status=.
func (s *Server) handleApprovalList(w http.ResponseWriter, r *http.Request) {
	approvals, err := s.db.ListApprovals(r.URL.Query().Get("status"))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "list_failed", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, approvals)
}

// handleApprovalGet returns one approval.
func (s *Server) handleApprovalGet(w http.ResponseWriter, r *http.Request) {
	a, err := s.db.GetApproval(r.PathValue("id"))
	if err != nil {
		s.writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, a)
}

func (s *Server) handleApprovalApprove(w http.ResponseWriter, r *http.Request) {
	s.decideApproval(w, r, approval.StatusApproved)
}

func (s *Server) handleApprovalReject(w http.ResponseWriter, r *http.Request) {
	s.decideApproval(w, r, approval.StatusRejected)
}

func (s *Server) handleApprovalClarify(w http.ResponseWriter, r *http.Request) {
	s.decideApproval(w, r, approval.StatusClarifying)
}

// decideApproval records a decision via the approval service (JSON API path).
func (s *Server) decideApproval(w http.ResponseWriter, r *http.Request, to approval.Status) {
	var body struct {
		Reason string `json:"reason"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)
	if a, ok := s.applyDecision(w, r.PathValue("id"), to, body.Reason); ok {
		writeJSON(w, http.StatusOK, a)
	}
}

// applyDecision records an approval decision through the ApprovalService and runs
// the gate's side effects, shared by the JSON API and the operator UI form so
// neither bypasses policy or self-approves (ADR 0028). On failure it writes an
// error response and returns ok=false; the caller writes the success response.
func (s *Server) applyDecision(w http.ResponseWriter, id string, to approval.Status, reason string) (*sqlite.Approval, bool) {
	if s.approvals == nil {
		writeError(w, http.StatusServiceUnavailable, "approvals_unavailable", "approval service is not configured")
		return nil, false
	}
	a, err := s.approvals.Decide(id, to, ownerActor, reason)
	if err != nil {
		s.writeMutationError(w, err)
		return nil, false
	}
	// Accepting a decomposition ratifies the parent contract into canon
	// (ADR 0013/0030); record the ratification gate in the node's journal.
	if to == approval.StatusApproved && approval.Type(a.Type) == approval.TypeDecompose {
		s.ratify(a.TicketID, "decompose", "decomposition accepted; parent contract promoted")
	}
	// Approving a land_to_main gate performs the land in Decide; complete it with
	// the durable git commit so the node is committed, not just recorded (ADR 0034).
	if to == approval.StatusApproved && approval.Type(a.Type) == approval.TypeLandToMain {
		if !s.completeLanding(w, a.TicketID, "node landed (human-approved)") {
			return nil, false
		}
	}
	return a, true
}

// handleTicketDecompose records a decomposition proposal: children in backlog +
// a pending decompose approval (work-tree.md, ADR 0030).
func (s *Server) handleTicketDecompose(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Contract json.RawMessage    `json:"contract"`
		Children []sqlite.ChildSpec `json:"children"`
	}
	if !decodeJSON(w, r, &body) {
		return
	}
	contract := "{}"
	if len(body.Contract) > 0 {
		contract = string(body.Contract)
	}
	appr, childIDs, err := s.db.DecomposeProposal(r.PathValue("id"), contract, body.Children, ownerActor)
	if err != nil {
		s.writeMutationError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"approval": appr, "child_ids": childIDs})
}

// handleTicketEscalate records a typed escalation and opens a re-plan approval.
func (s *Server) handleTicketEscalate(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Reason string `json:"reason"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)
	appr, err := s.db.Escalate(r.PathValue("id"), body.Reason, ownerActor)
	if err != nil {
		s.writeMutationError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, appr)
}

// ApprovalService creates approvals from gate decisions and records decisions
// with actor authorization, dispatching type-specific side effects (ADR 0028).
// It lives in the server package to avoid an approval<->store import cycle while
// keeping logic out of the HTTP handlers.
type ApprovalService struct {
	db       *sqlite.DB
	policies *policy.Set
	registry *actor.Registry
}

// NewApprovalService builds the service.
func NewApprovalService(db *sqlite.DB, policies *policy.Set, registry *actor.Registry) *ApprovalService {
	return &ApprovalService{db: db, policies: policies, registry: registry}
}

// RequestParams describes a gated action seeking approval.
type RequestParams struct {
	RunID          string
	TicketID       string
	Type           approval.Type
	Summary        string
	Action         policy.Action
	ActionJSON     string
	RequiredActors []string
	RequiredRoles  []string
}

// Request evaluates the gate for an action and records an approval. When policy
// auto-approves (low-risk reversible docs, ADR 0028/T-0603) the approval is
// stored already-approved with the firing rule as the decider; otherwise it is
// recorded pending for a human.
func (s *ApprovalService) Request(p RequestParams) (*sqlite.Approval, error) {
	if !p.Type.Valid() {
		return nil, fmt.Errorf("invalid approval type %q", p.Type)
	}
	d := s.policies.Evaluate(p.Action)
	score, reversible := d.RiskScore, d.Reversible

	params := sqlite.CreateApprovalParams{
		RunID: p.RunID, TicketID: p.TicketID, Type: p.Type, RiskClass: string(d.RiskClass),
		RiskScore: &score, Reversible: &reversible, Summary: p.Summary, ActionJSON: p.ActionJSON,
		RequestedByActor: actorID(p.Action.Actor), RequiredActors: p.RequiredActors, RequiredRoles: p.RequiredRoles,
	}
	if d.Outcome == policy.OutcomeAutoApprove {
		params.Status = approval.StatusApproved
		params.DecidedByActor = "policy:" + d.RuleID
		params.DecisionReason = "auto-approved by policy"
	} else {
		params.Status = approval.StatusPending
	}
	return s.db.CreateApproval(params)
}

// RequestLanding opens a land_to_main approval for a node through the gate engine
// (ADR 0028): low-risk reversible changes the policy auto-approves come back
// already-approved; everything else is pending for a human. The owner is the
// requesting actor in single-user v1. The changed-file Scope is empty in M2 (it
// arrives with the Phase 4 runtime's diff), so in practice landing is
// human-gated until then.
func (s *ApprovalService) RequestLanding(ticketID, workType string) (*sqlite.Approval, error) {
	owner, _ := s.registry.Resolve(ownerActor)
	return s.Request(RequestParams{
		TicketID: ticketID,
		Type:     approval.TypeLandToMain,
		Summary:  "Land " + ticketID,
		Action:   policy.Action{Type: "land_to_main", WorkType: workType, Actor: owner},
	})
}

// Decide authorizes the deciding actor against the approval's required
// actors/roles, then records the decision and dispatches type-specific side
// effects (decompose creates children; replan requeues the node). require_human
// is never bypassed: auto-approved approvals never reach here.
func (s *ApprovalService) Decide(id string, to approval.Status, decidedBy, reason string) (*sqlite.Approval, error) {
	a, err := s.db.GetApproval(id)
	if err != nil {
		return nil, err
	}
	if err := s.authorize(a, decidedBy); err != nil {
		return nil, err
	}
	switch to {
	case approval.StatusApproved:
		switch approval.Type(a.Type) {
		case approval.TypeDecompose:
			return s.db.AcceptDecompose(id, decidedBy, reason)
		case approval.TypeReplan:
			return s.db.AcceptReplan(id, decidedBy, reason)
		case approval.TypeLandToMain:
			// Approving the landing performs the validated land. The validation
			// gate is enforced here; if it blocks, the approval stays pending so a
			// human can fix validation (or land --override). M2 supplies no
			// changed-file set, so required checks are empty and the gate enforces
			// "no failing results".
			if _, err := s.db.Land(a.TicketID, nil, false, decidedBy); err != nil {
				return nil, err
			}
			return s.db.DecideApproval(id, approval.StatusApproved, decidedBy, reason)
		default:
			return s.db.DecideApproval(id, approval.StatusApproved, decidedBy, reason)
		}
	case approval.StatusRejected:
		switch approval.Type(a.Type) {
		case approval.TypeDecompose:
			return s.db.RejectDecompose(id, decidedBy, reason)
		case approval.TypeReplan:
			return s.db.RejectReplan(id, decidedBy, reason)
		default:
			return s.db.DecideApproval(id, approval.StatusRejected, decidedBy, reason)
		}
	case approval.StatusClarifying:
		return s.db.DecideApproval(id, approval.StatusClarifying, decidedBy, reason)
	default:
		return nil, fmt.Errorf("invalid decision %q", to)
	}
}

// authorize enforces the approval's actor/role constraints against the deciding
// actor (ADR 0029 prefix/role matching).
func (s *ApprovalService) authorize(a *sqlite.Approval, decidedBy string) error {
	dec, ok := s.registry.Resolve(decidedBy)
	if !ok {
		return fmt.Errorf("deciding actor %q is not in the registry", decidedBy)
	}
	if len(a.RequiredActors) > 0 && !actor.AnyIDMatch(a.RequiredActors, dec.ID) {
		return fmt.Errorf("actor %q is not permitted to decide approval %s", dec.ID, a.ID)
	}
	if len(a.RequiredRoles) > 0 && !dec.HasAnyRole(a.RequiredRoles) {
		return fmt.Errorf("actor %q lacks a required role to decide approval %s", dec.ID, a.ID)
	}
	// v1 human gate: with no explicit actor/role constraints, a human-gated action
	// may only be decided by a human actor. This keeps require_human un-bypassable
	// at the record level when future AI chat/reviewer adapters call Decide
	// (ADR 0028). Auto-approved gates never reach Decide.
	if len(a.RequiredActors) == 0 && len(a.RequiredRoles) == 0 && humanGated(approval.Type(a.Type)) {
		if dec.Type != actor.TypeHuman {
			return fmt.Errorf("approval %s is human-gated; actor %q (%s) may not decide it", a.ID, dec.ID, dec.Type)
		}
	}
	return nil
}

// humanGated reports whether an approval type is human-required in v1.
func humanGated(t approval.Type) bool {
	switch t {
	case approval.TypeDecompose, approval.TypeReplan, approval.TypeLandToMain:
		return true
	}
	return false
}

func actorID(a *actor.Actor) string {
	if a == nil {
		return ""
	}
	return a.ID
}
