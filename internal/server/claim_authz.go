package server

// Envelope-aware claim authorization (ADR 0056): an AI action is authorized only
// when trust policy AND an active envelope both permit it (AND-composition). No
// active envelope denies AI claims (default-deny, identical to today). A boundary
// crossing — trust would allow the action but the envelope would not — raises a
// human exception approval instead of silently failing. Humans bypass envelopes.

import (
	"fmt"

	"groundwork/internal/actor"
	"groundwork/internal/approval"
	"groundwork/internal/envelope"
	"groundwork/internal/policy"
	"groundwork/internal/risk"
	"groundwork/internal/store/sqlite"
)

// ClaimOutcome is the result of envelope-aware claim authorization.
type ClaimOutcome string

const (
	ClaimAllow     ClaimOutcome = "allow"
	ClaimDeny      ClaimOutcome = "deny"
	ClaimException ClaimOutcome = "exception" // raised a human exception approval
)

// AuthorizeEnvelopedClaim composes trust policy with the active ancestor envelope
// for an AI action on nodeID (ADR 0056). It returns the outcome and, for a
// boundary crossing, the exception approval it opened.
func (s *Server) AuthorizeEnvelopedClaim(nodeID, action, workType string, a *actor.Actor, class risk.Class, files []string) (ClaimOutcome, *sqlite.Approval, error) {
	// Humans may always pick up their own work; envelopes bound AI autonomy.
	if a != nil && a.Type == actor.TypeHuman {
		return ClaimAllow, nil, nil
	}
	envID, within, planned, err := s.envelopeFacts(nodeID, envelopeActionFor(action), firstRole(a), workType, class, files)
	if err != nil {
		return ClaimDeny, nil, err
	}
	if envID == "" {
		// No envelope governs this node: AI claims are not authorized.
		return ClaimDeny, nil, nil
	}
	// Would trust authorize this action if it were inside the envelope? This
	// isolates the envelope as the only blocker, so a crossing becomes an exception
	// rather than a plain deny.
	probe := policy.Action{
		Type: action, Actor: a, WorkType: workType,
		Scope:          risk.Scope{Files: files},
		ActorRole:      firstRole(a),
		EnvelopeID:     envID,
		PlannedScope:   planned,
		WithinEnvelope: true,
	}
	trustAllows := s.approvals.policies.AuthorizeClaim(probe).Outcome == policy.OutcomeAllow
	if !trustAllows {
		return ClaimDeny, nil, nil
	}
	if within {
		return ClaimAllow, nil, nil
	}
	// Dedup: a boundary-crossing node is re-evaluated on every scheduler tick. Only
	// raise an exception when one is not already open for it, so the queue and the
	// durable sidecar are not flooded with duplicate requests.
	if open, err := s.db.HasOpenApprovalOfType(nodeID, approval.TypeException); err != nil {
		return ClaimDeny, nil, err
	} else if open {
		return ClaimException, nil, nil
	}
	appr, err := s.raiseEnvelopeException(nodeID, envID, action, a)
	if err != nil {
		return ClaimDeny, nil, err
	}
	return ClaimException, appr, nil
}

// AuthorizeAIClaim is the scheduler-facing entry point (ADR 0056): it computes
// the claim-time risk class for the action and runs the envelope-aware claim
// composition with no diff yet (planned scope governs at claim; actual scope is
// enforced once the runtime supplies the diff, T-1091). It returns the outcome.
func (s *Server) AuthorizeAIClaim(nodeID, action, workType string, a *actor.Actor) (ClaimOutcome, error) {
	class := risk.ClassLow
	if s.approvals != nil {
		class = s.approvals.policies.Evaluate(policy.Action{Type: action, Actor: a, WorkType: workType}).RiskClass
	}
	outcome, _, err := s.AuthorizeEnvelopedClaim(nodeID, action, workType, a, class, nil)
	return outcome, err
}

// raiseEnvelopeException opens a human-gated exception approval for an action that
// exceeds its envelope, linked to the node and the governing envelope.
func (s *Server) raiseEnvelopeException(nodeID, envID, action string, a *actor.Actor) (*sqlite.Approval, error) {
	return s.approvals.Request(RequestParams{
		TicketID:   nodeID,
		Type:       approval.TypeException,
		Summary:    fmt.Sprintf("Exception: %s on %s exceeds envelope %s", action, nodeID, envID),
		Action:     policy.Action{Type: string(approval.TypeException), Actor: a},
		ActionJSON: fmt.Sprintf(`{"envelope_id":%q,"action":%q}`, envID, action),
	})
}

// envelopeActionFor maps a gate action (execute/decompose/land_to_parent/replan)
// to the envelope's approved-action vocabulary (ADR 0054/0056). An unrecognized
// gate action maps to "" — no envelope grant matches it, so it is never
// authorized by an envelope. The raw gate action is deliberately NOT passed
// through, so the gate and envelope vocabularies stay decoupled and a future gate
// action can never silently match an envelope grant by name collision.
func envelopeActionFor(gateAction string) string {
	switch gateAction {
	case "execute":
		return envelope.ActionExecuteChildren
	case "decompose":
		return envelope.ActionDecomposeChildren
	case "land_to_parent":
		return envelope.ActionLandChildToParent
	case "replan":
		return envelope.ActionReplanWithinGoal
	}
	return ""
}

// firstRole returns an actor's primary role, or "" (the v1 AI role actors hold a
// single role; the human owner bypasses envelopes before this is reached).
func firstRole(a *actor.Actor) string {
	if a != nil && len(a.Roles) > 0 {
		return a.Roles[0]
	}
	return ""
}
