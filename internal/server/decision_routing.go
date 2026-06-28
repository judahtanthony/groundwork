package server

// Consequential-decision routing (ADR 0052): a blocked run either raises a real
// decision work node (independent scope/ownership/canon impact) or records a
// bounded local input request. Both produce durable ticket-attached records so
// the block survives a rebuild; only the consequential branch spawns a ticket.

import (
	"net/http"

	"groundwork/internal/resume"
	"groundwork/internal/store/sqlite"
)

// handleTicketResume returns the durable resume packet for a node (ADR 0051): the
// structured context a new run starts from instead of a live session.
func (s *Server) handleTicketResume(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if _, err := s.db.GetTicket(id); err != nil {
		s.writeStoreError(w, err)
		return
	}
	packet, err := resume.Assemble(s.db, id)
	if err != nil {
		s.writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, packet)
}

// handleTicketRaiseDecision creates a decision work node, links the blocked
// ticket to it, and records a durable decision_requested record (ADR 0052).
func (s *Server) handleTicketRaiseDecision(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Title          string   `json:"title"`
		WorkType       string   `json:"work_type"`
		RequestedActor string   `json:"requested_actor"`
		Statement      string   `json:"statement"`
		Acceptance     []string `json:"acceptance"`
		RunID          string   `json:"run_id"`
		RequestedBy    string   `json:"requested_by"`
		Parent         string   `json:"parent"`
	}
	if !decodeJSON(w, r, &body) {
		return
	}
	id := r.PathValue("id")
	decisionID, rec, err := s.db.RaiseDecision(sqlite.RaiseDecisionParams{
		BlockedTicketID: id, RunID: body.RunID, Title: body.Title, WorkType: body.WorkType,
		RequestedActor: body.RequestedActor, Statement: body.Statement, Acceptance: body.Acceptance,
		RequestedBy: body.RequestedBy, Parent: body.Parent,
	})
	if err != nil {
		s.writeMutationError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{
		"blocked_ticket": id, "decision_ticket": decisionID, "record": rec,
	})
}

// handleTicketRequestInput records a bounded local input request without
// creating a work node (ADR 0052).
func (s *Server) handleTicketRequestInput(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Statement   string `json:"statement"`
		RunID       string `json:"run_id"`
		RequestedBy string `json:"requested_by"`
	}
	if !decodeJSON(w, r, &body) {
		return
	}
	id := r.PathValue("id")
	rec, err := s.db.RequestInput(id, body.RunID, body.Statement, body.RequestedBy)
	if err != nil {
		s.writeMutationError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"ticket": id, "record": rec})
}
