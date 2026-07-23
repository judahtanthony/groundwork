package server

import (
	"net/http"

	"groundwork/internal/contextbrief"
	"groundwork/internal/ticket"
)

type readinessBlocker struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

type blockedReadinessNode struct {
	*ticket.Ticket
	BlockedBy []readinessBlocker `json:"blocked_by"`
}

type recommendedReadinessNode struct {
	Ticket *ticket.Ticket      `json:"ticket"`
	Brief  *contextbrief.Brief `json:"brief"`
}

type readinessResponse struct {
	Next    *recommendedReadinessNode `json:"next"`
	Ready   []*ticket.Ticket          `json:"ready"`
	Blocked []blockedReadinessNode    `json:"blocked"`
}

// handleReadiness returns the same operator-facing readiness view as gw next,
// gw ticket list --ready, and gw ticket list --blocked. Value ordering stays in
// the store so the CLI, scheduler, API, and SPA cannot disagree.
func (s *Server) handleReadiness(w http.ResponseWriter, _ *http.Request) {
	ready, err := s.db.ListEligibleOrdered()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "list_failed", err.Error())
		return
	}
	if ready == nil {
		ready = []*ticket.Ticket{}
	}

	all, err := s.db.ListTickets()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "list_failed", err.Error())
		return
	}
	blocked := []blockedReadinessNode{}
	for _, node := range all {
		if node.Status != ticket.StatusTodo {
			continue
		}
		depIDs, err := s.db.DependencyIDs(node.ID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "list_failed", err.Error())
			return
		}
		blockers := []readinessBlocker{}
		for _, depID := range depIDs {
			dep, err := s.db.GetTicket(depID)
			if err != nil {
				s.writeStoreError(w, err)
				return
			}
			if !ticket.DependencyMet(dep.Status) {
				blockers = append(blockers, readinessBlocker{ID: dep.ID, Status: string(dep.Status)})
			}
		}
		if len(blockers) > 0 {
			blocked = append(blocked, blockedReadinessNode{Ticket: node, BlockedBy: blockers})
		}
	}

	var next *recommendedReadinessNode
	if len(ready) > 0 {
		brief, err := contextbrief.Build(s.db, s.proj, ready[0].ID, false)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "context_failed", err.Error())
			return
		}
		next = &recommendedReadinessNode{Ticket: ready[0], Brief: brief}
	}

	writeJSON(w, http.StatusOK, readinessResponse{Next: next, Ready: ready, Blocked: blocked})
}
