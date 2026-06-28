package server

// Envelope facts for the gate (ADR 0056): the coordinator resolves the active
// ancestor envelope for a node and computes whether an action sits within the
// approved boundary (action, work type, role, scope, risk), so policy can match
// on within_envelope and so boundary crossings can raise exceptions.

import (
	"groundwork/internal/envelope"
	"groundwork/internal/policy"
	"groundwork/internal/risk"
)

// activeAncestorEnvelope returns the active envelope nearest to nodeID (self,
// then closest ancestor), or (nil, nil) when none is active in the chain.
func (s *Server) activeAncestorEnvelope(nodeID string) (*envelope.Envelope, error) {
	return nearestInChain(s, nodeID, s.db.GetActiveEnvelopeForNode)
}

// envelopeFacts computes the ADR 0056 gate inputs for an action on nodeID: the
// governing envelope id, whether the action is within it, and its planned scope.
// within is false (and id empty) when no active envelope governs the node.
func (s *Server) envelopeFacts(nodeID, action, actorRole, workType string, class risk.Class, files []string) (envID string, within bool, planned []string, err error) {
	e, err := s.activeAncestorEnvelope(nodeID)
	if err != nil || e == nil {
		return "", false, nil, err
	}
	return e.ID, envelopeAuthorizes(e, action, actorRole, workType, class, files), e.Scope.Files.Allow, nil
}

// envelopeAuthorizes reports whether an action sits fully within the envelope:
// approved action, allowed work type, allowed role, risk at or below ceiling, and
// file scope within allow / not in deny (ADR 0056).
func envelopeAuthorizes(e *envelope.Envelope, action, actorRole, workType string, class risk.Class, files []string) bool {
	if !e.Allows(action) {
		return false
	}
	if workType != "" && !e.AllowsWorkType(workType) {
		return false
	}
	if actorRole != "" && !e.AllowsRole(actorRole) {
		return false
	}
	if e.RiskCeiling != "" && !class.AtMost(risk.Class(e.RiskCeiling)) {
		return false
	}
	return envelopeScopeAllows(e, files)
}

// envelopeScopeAllows enforces the file scope: nothing in deny, and (when an
// allow-list is set) every touched file within it.
func envelopeScopeAllows(e *envelope.Envelope, files []string) bool {
	if len(files) == 0 {
		return true
	}
	if len(e.Scope.Files.Deny) > 0 && policy.FilesMatch(e.Scope.Files.Deny, files) {
		return false
	}
	if len(e.Scope.Files.Allow) == 0 {
		return true
	}
	for _, f := range files {
		if !policy.FilesMatch(e.Scope.Files.Allow, []string{f}) {
			return false
		}
	}
	return true
}
