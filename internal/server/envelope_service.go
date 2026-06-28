package server

// Envelope lifecycle (ADR 0054): propose → human-approve (activate) → revoke /
// supersede. Proposal carries the draft envelope in the approval's action JSON;
// the envelope record (sidecar + mirror) is created only on approval, so an
// unapproved boundary never authorizes anything. Child creation is the existing
// decompose flow, which composes within an approved envelope.

import (
	"encoding/json"
	"fmt"

	"groundwork/internal/approval"
	"groundwork/internal/encoding"
	"groundwork/internal/envelope"
	"groundwork/internal/policy"
	"groundwork/internal/store/sqlite"
)

// ProposeEnvelope opens a human-gated approve_envelope approval for nodeID,
// carrying the draft boundary. Activation happens on approval (recordDecision).
func (s *Server) ProposeEnvelope(nodeID string, draft *envelope.Envelope) (*sqlite.Approval, error) {
	if s.approvals == nil {
		return nil, errApprovalsUnavailable
	}
	if _, err := s.db.GetTicket(nodeID); err != nil {
		return nil, err
	}
	draft.NodeID = nodeID
	draftJSON, err := json.Marshal(draft)
	if err != nil {
		return nil, err
	}
	owner, _ := s.approvals.registry.Resolve(ownerActor)
	return s.approvals.Request(RequestParams{
		TicketID:   nodeID,
		Type:       approval.TypeApproveEnvelope,
		Summary:    "Approve envelope for " + nodeID,
		Action:     policy.Action{Type: string(approval.TypeApproveEnvelope), Actor: owner},
		ActionJSON: string(draftJSON),
	})
}

// activateEnvelope is invoked when an approve_envelope approval is approved: it
// materializes the draft (from action JSON) as an active envelope — authoritative
// sidecar plus SQLite mirror.
func (s *Server) activateEnvelope(actionJSON, nodeID, decidedBy string) error {
	var draft envelope.Envelope
	if err := json.Unmarshal([]byte(actionJSON), &draft); err != nil {
		return fmt.Errorf("envelope draft: %w", err)
	}
	id, err := s.db.NextEnvelopeID()
	if err != nil {
		return err
	}
	draft.ID = id
	draft.NodeID = nodeID
	draft.Status = envelope.StatusActive
	draft.ApprovedBy = decidedBy
	draft.ApprovedAt = encoding.Now()
	if err := envelope.Write(s.proj.TicketsDir(), &draft); err != nil {
		return err
	}
	if err := s.db.UpsertEnvelope(&draft); err != nil {
		return err
	}
	// Establish the root's integration target so children can land to it (ADR 0058).
	return s.ensureIntegrationBranch(nodeID)
}

// RevokeEnvelope flips an envelope to revoked in both the mirror and the
// authoritative sidecar; new claims/landings under it are then refused (ADR 0054).
func (s *Server) RevokeEnvelope(id string) error {
	return s.setEnvelopeStatus(id, envelope.StatusRevoked)
}

// SupersedeEnvelope marks an envelope superseded (a re-plan replaces it).
func (s *Server) SupersedeEnvelope(id string) error {
	return s.setEnvelopeStatus(id, envelope.StatusSuperseded)
}

func (s *Server) setEnvelopeStatus(id string, status envelope.Status) error {
	e, err := s.db.GetEnvelope(id)
	if err != nil {
		return err
	}
	// Write the authoritative sidecar before the SQLite mirror, matching
	// activateEnvelope's order (ADR 0040/0054): the file is the source of truth, so
	// if the mirror write fails the sidecar already reflects the new status and a
	// rebuild reconciles the mirror — never the reverse, which would leave the
	// authoritative file stale behind a changed mirror.
	e.Status = status
	if err := envelope.Write(s.proj.TicketsDir(), e); err != nil {
		return err
	}
	return s.db.SetEnvelopeStatus(id, status)
}
