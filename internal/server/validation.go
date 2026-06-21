package server

import (
	"encoding/json"
	"net/http"

	"groundwork/internal/approval"
	"groundwork/internal/canon"
	"groundwork/internal/eventbus"
	"groundwork/internal/store/sqlite"
	"groundwork/internal/ticket"
)

// handleTicketValidations returns a node's recorded validation results (T-0702).
func (s *Server) handleTicketValidations(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if _, err := s.db.GetTicket(id); err != nil {
		s.writeStoreError(w, err)
		return
	}
	results, err := s.db.ListValidationsForTicket(id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "list_failed", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, results)
}

// handleRecordValidation records a validation result through the coordinator so
// the running server's state and SSE stream stay coherent (ADR 0031): the
// `gw validation run` CLI executes the checks locally and posts each outcome here
// rather than writing the store directly.
func (s *Server) handleRecordValidation(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if _, err := s.db.GetTicket(id); err != nil {
		s.writeStoreError(w, err)
		return
	}
	var body struct {
		Name         string `json:"name"`
		Command      string `json:"command"`
		Status       string `json:"status"`
		ArtifactPath string `json:"artifact_path"`
	}
	if !decodeJSON(w, r, &body) {
		return
	}
	res, err := s.db.RecordValidation(sqlite.ValidationResult{
		TicketID: id, Name: body.Name, Command: body.Command, Status: body.Status, ArtifactPath: body.ArtifactPath,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "record_failed", err.Error())
		return
	}
	if s.bus != nil {
		s.bus.Publish(eventbus.Event{Type: "validation.recorded", TicketID: id, Message: body.Name + ": " + body.Status})
	}
	writeJSON(w, http.StatusCreated, res)
}

// landResponse is the POST /land result: either the node was landed (auto-
// approved or override), or a human approval is now pending.
type landResponse struct {
	Landed   bool             `json:"landed"`
	Ticket   *ticket.Ticket   `json:"ticket,omitempty"`
	Approval *sqlite.Approval `json:"approval,omitempty"`
}

// handleTicketLand drives landing through the land_to_main approval gate
// (ADR 0028/0006). Normal path: open a land_to_main approval — if policy
// auto-approves, land now; otherwise return the pending approval for a human to
// decide (approving it lands). The `override` escape hatch lets the owner land
// immediately, bypassing both the approval gate and the validation gate. A
// successful landing records a ratification in the journal (ADR 0013).
func (s *Server) handleTicketLand(w http.ResponseWriter, r *http.Request) {
	if s.approvals == nil {
		writeError(w, http.StatusServiceUnavailable, "approvals_unavailable", "approval service is not configured")
		return
	}
	var body struct {
		Override bool `json:"override"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body) // empty body is allowed
	id := r.PathValue("id")

	node, err := s.db.GetTicket(id)
	if err != nil {
		s.writeStoreError(w, err)
		return
	}

	if body.Override {
		if _, err := s.db.Land(id, nil, true, ownerActor); err != nil {
			s.writeMutationError(w, err)
			return
		}
		s.ratify(id, "land", "node landed (override)")
		if err := s.commitLanding(id); err != nil {
			writeError(w, http.StatusInternalServerError, "land_commit_failed", err.Error())
			return
		}
		s.landed(w, id)
		return
	}

	appr, err := s.approvals.RequestLanding(id, node.WorkType)
	if err != nil {
		s.writeMutationError(w, err)
		return
	}
	if approvalApproved(appr) { // policy auto-approved
		if _, err := s.db.Land(id, nil, false, ownerActor); err != nil {
			s.writeMutationError(w, err)
			return
		}
		s.ratify(id, "land", "node landed (auto-approved by policy)")
		if err := s.commitLanding(id); err != nil {
			writeError(w, http.StatusInternalServerError, "land_commit_failed", err.Error())
			return
		}
		s.landed(w, id)
		return
	}
	writeJSON(w, http.StatusOK, landResponse{Landed: false, Approval: appr})
}

// landed writes a "landed" response with the updated node.
func (s *Server) landed(w http.ResponseWriter, id string) {
	t, err := s.db.GetTicket(id)
	if err != nil {
		s.writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, landResponse{Landed: true, Ticket: t})
}

func approvalApproved(a *sqlite.Approval) bool {
	return a != nil && approval.Status(a.Status) == approval.StatusApproved
}

// ratify records a ratification-gate event in the node's journal (best-effort;
// the canon write is serialized through this single coordinator, ADR 0013/0030).
func (s *Server) ratify(nodeID, gate, note string) {
	if s.proj == nil {
		return
	}
	_ = canon.Ratify(s.proj.JournalDir(), nodeID, gate, note)
}
