package server

// Tickets screen (T-1063): the operator's read of the work tree — a value-ordered
// ready queue, a blocked queue annotated with unmet dependencies, and the full
// ticket list with each todo node's ready/blocked state. It reuses the same store
// reads the CLI uses (gw ticket list --ready/--blocked) so the surfaces never
// diverge (ADR 0041): ready is ListEligibleOrdered (ADR 0039 value order) and
// blocked mirrors listBlocked's unmet-dependency annotation.

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"groundwork/internal/ticket"
)

var ticketsTmpl = newPage("web/tickets.content.tmpl")

// ticketRow is a ticket as shown in the ready and all-tickets tables. State is
// "ready"/"blocked" for todo nodes and "" otherwise.
type ticketRow struct {
	ID, Title, Status, Priority, NodeType, WorkType, Parent, State string
}

// blockedRow is a blocked todo node plus a rendered list of its unmet deps.
type blockedRow struct {
	ID, Title, NodeType, Parent, BlockedBy string
}

type ticketsData struct {
	Ready   []ticketRow
	Blocked []blockedRow
	All     []ticketRow
}

func (s *Server) handleTicketsPage(w http.ResponseWriter, r *http.Request) {
	data, err := s.buildTickets()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "tickets_failed", err.Error())
		return
	}
	s.renderPage(w, ticketsTmpl, &pageView{
		Shell: s.shellState(s.pendingCount()),
		Nav:   navTickets,
		Crumb: "Operate",
		Title: "Tickets",
		Data:  data,
	})
}

func (s *Server) buildTickets() (*ticketsData, error) {
	all, err := s.db.ListTickets()
	if err != nil {
		return nil, err
	}
	ready, err := s.db.ListEligibleOrdered()
	if err != nil {
		return nil, err
	}

	d := &ticketsData{}
	readyIDs := make(map[string]bool, len(ready))
	for _, t := range ready {
		readyIDs[t.ID] = true
		d.Ready = append(d.Ready, ticketRowOf(t, "ready"))
	}

	// Blocked: todo nodes with at least one unmet dependency, annotated with the
	// blocking deps and their statuses (mirrors gw ticket list --blocked).
	blockedIDs := map[string]bool{}
	for _, t := range all {
		if t.Status != ticket.StatusTodo {
			continue
		}
		depIDs, err := s.db.DependencyIDs(t.ID)
		if err != nil {
			return nil, err
		}
		var parts []string
		for _, depID := range depIDs {
			dep, err := s.db.GetTicket(depID)
			if err != nil {
				return nil, err
			}
			if !ticket.DependencyMet(dep.Status) {
				parts = append(parts, fmt.Sprintf("%s (%s)", dep.ID, dep.Status))
			}
		}
		if len(parts) > 0 {
			blockedIDs[t.ID] = true
			d.Blocked = append(d.Blocked, blockedRow{
				ID: t.ID, Title: t.Title,
				NodeType: orDash(string(t.NodeType)), Parent: orDash(t.ParentID),
				BlockedBy: strings.Join(parts, ", "),
			})
		}
	}

	for _, t := range all {
		state := ""
		if t.Status == ticket.StatusTodo {
			switch {
			case readyIDs[t.ID]:
				state = "ready"
			case blockedIDs[t.ID]:
				state = "blocked"
			}
		}
		d.All = append(d.All, ticketRowOf(t, state))
	}
	return d, nil
}

func ticketRowOf(t *ticket.Ticket, state string) ticketRow {
	return ticketRow{
		ID:       t.ID,
		Title:    t.Title,
		Status:   string(t.Status),
		Priority: strconv.FormatFloat(t.EffectivePriority(), 'f', 2, 64),
		NodeType: orDash(string(t.NodeType)),
		WorkType: orDash(t.WorkType),
		Parent:   orDash(t.ParentID),
		State:    state,
	}
}
