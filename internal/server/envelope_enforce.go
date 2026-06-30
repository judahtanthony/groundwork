package server

// Diff-fed envelope enforcement (ADR 0056, T-1091): once the runtime supplies a
// node's changed-file set (ADR 0059), the coordinator enforces the envelope's
// file scope and escalation triggers at review/landing. A violation raises a
// human exception approval and blocks the landing — the human-gated invariants
// (unexpected scope, contract/public-API change, failed validation, risk above
// the ceiling) are never silently bypassed.

import (
	"errors"
	"fmt"
	"strings"

	"groundwork/internal/approval"
	"groundwork/internal/envelope"
	"groundwork/internal/policy"
	"groundwork/internal/risk"
	"groundwork/internal/store/sqlite"
)

// ErrEnvelopeEscalation is returned when diff-fed enforcement raises an exception
// and blocks the action pending human approval.
var ErrEnvelopeEscalation = errors.New("envelope escalation: action raised a human exception")

// enforceEnvelopeOnDiff evaluates the active envelope's file scope and escalation
// triggers against the node's captured changed-file set (ADR 0056). When any
// fires it opens a single exception approval listing the reasons and returns
// ErrEnvelopeEscalation. With no governing envelope or no triggered rule it is a
// no-op (nil, nil), so manual single-tree work and ungoverned nodes are unaffected.
func (s *Server) enforceEnvelopeOnDiff(nodeID, action string) (*sqlite.Approval, error) {
	env, err := s.activeAncestorEnvelope(nodeID)
	if err != nil || env == nil {
		return nil, err
	}
	files, err := s.db.ChangedFilesForNode(nodeID)
	if err != nil {
		return nil, err
	}
	reasons := s.envelopeViolations(env, nodeID, action, files)
	if len(reasons) == 0 {
		return nil, nil
	}
	// Dedup: don't stack duplicate exceptions when a blocked landing is retried.
	if open, oerr := s.db.HasOpenApprovalOfType(nodeID, approval.TypeException); oerr != nil {
		return nil, oerr
	} else if open {
		return nil, ErrEnvelopeEscalation
	}
	appr, rerr := s.raiseEscalation(nodeID, env.ID, action, reasons)
	if rerr != nil {
		return nil, rerr
	}
	return appr, ErrEnvelopeEscalation
}

// envelopeViolations returns the human-readable reasons an action breaches its
// envelope given the diff (ADR 0056). Each escalation trigger is only evaluated
// when enabled on the envelope.
func (s *Server) envelopeViolations(env *envelope.Envelope, nodeID, action string, files []string) []string {
	var reasons []string

	// Actual-vs-planned file scope (ADR 0046): a touched file outside allow or in
	// deny is an unexpected scope expansion.
	if len(files) > 0 && !envelopeScopeAllows(env, files) {
		reasons = append(reasons, "changed files fall outside the envelope's file scope")
	}
	// require_review paths always need a human look, regardless of triggers.
	if len(env.Scope.Files.RequireReview) > 0 && policy.FilesMatch(env.Scope.Files.RequireReview, files) {
		reasons = append(reasons, "changed files match a require_review path")
	}

	esc := env.Escalation
	if esc.OnUnexpectedFiles && hasUnexpectedFiles(env, files) {
		reasons = append(reasons, "unexpected files outside the envelope allow-list")
	}
	if esc.OnContractChange && touchesContracts(files) {
		reasons = append(reasons, "a contract document changed")
	}
	if esc.OnPublicAPIChange && touchesPublicAPI(files) {
		reasons = append(reasons, "a public API surface changed")
	}
	if esc.OnValidationFailure && s.hasFailingValidation(nodeID) {
		reasons = append(reasons, "a validation result is failing")
	}
	if esc.OnRiskAboveCeiling && s.riskAboveCeiling(env, action, files) {
		reasons = append(reasons, "the change risk exceeds the envelope ceiling")
	}
	return reasons
}

// hasUnexpectedFiles reports whether any touched file is outside the envelope's
// allow-list (an empty allow-list means "anything", so nothing is unexpected).
func hasUnexpectedFiles(env *envelope.Envelope, files []string) bool {
	if len(env.Scope.Files.Allow) == 0 {
		return false
	}
	for _, f := range files {
		if !policy.FilesMatch(env.Scope.Files.Allow, []string{f}) {
			return true
		}
	}
	return false
}

// touchesContracts reports whether any changed file is canon contract material.
func touchesContracts(files []string) bool {
	for _, f := range files {
		if strings.HasPrefix(f, "docs/contracts/") || strings.HasPrefix(f, "docs/adr/") {
			return true
		}
	}
	return false
}

// touchesPublicAPI reports whether any changed file looks like a public API
// surface (heuristic v1): an api/ path, a protobuf/OpenAPI schema, or an exported
// HTTP route file. Tightened later as the resource model grows (ADR 0046).
func touchesPublicAPI(files []string) bool {
	for _, f := range files {
		l := strings.ToLower(f)
		if strings.Contains(l, "/api/") || strings.HasPrefix(l, "api/") ||
			strings.HasSuffix(l, ".proto") || strings.Contains(l, "openapi") {
			return true
		}
	}
	return false
}

// hasFailingValidation reports whether the node has any failing validation result.
func (s *Server) hasFailingValidation(nodeID string) bool {
	vals, err := s.db.ListValidationsForTicket(nodeID)
	if err != nil {
		return false
	}
	for _, v := range vals {
		if v.Status == sqlite.ValidationFail {
			return true
		}
	}
	return false
}

// riskAboveCeiling reports whether the change's risk class exceeds the envelope's
// ceiling, computed from the actual changed-file scope (ADR 0056).
func (s *Server) riskAboveCeiling(env *envelope.Envelope, action string, files []string) bool {
	if env.RiskCeiling == "" || s.approvals == nil {
		return false
	}
	class := s.approvals.policies.Evaluate(policy.Action{
		Type: action, Scope: risk.Scope{Files: files},
	}).RiskClass
	return !class.AtMost(risk.Class(env.RiskCeiling))
}

// raiseEscalation opens one human-gated exception approval carrying the triggered
// reasons, linked to the node and governing envelope.
func (s *Server) raiseEscalation(nodeID, envID, action string, reasons []string) (*sqlite.Approval, error) {
	summary := fmt.Sprintf("Exception: %s on %s — %s", action, nodeID, strings.Join(reasons, "; "))
	return s.approvals.Request(RequestParams{
		TicketID:   nodeID,
		Type:       approval.TypeException,
		Summary:    summary,
		Action:     policy.Action{Type: string(approval.TypeException)},
		ActionJSON: fmt.Sprintf(`{"envelope_id":%q,"action":%q,"reasons":%d}`, envID, action, len(reasons)),
	})
}
