package server

import (
	"net/http"
	"testing"

	"groundwork/internal/ticket"
)

func TestReadinessReturnsNextReadyAndBlocked(t *testing.T) {
	srv, db := newTestServer(t)
	low := mustCreate(t, db, &ticket.Ticket{Title: "lower value", Status: ticket.StatusTodo, Priority: priority(0.2)})
	high := mustCreate(t, db, &ticket.Ticket{
		Title: "higher value", Status: ticket.StatusTodo, Priority: priority(0.9),
		Acceptance: []string{"ship the ready surface"},
	})
	blocker := mustCreate(t, db, &ticket.Ticket{Title: "unfinished dependency", Status: ticket.StatusInProgress})
	blocked := mustCreate(t, db, &ticket.Ticket{Title: "waiting", Status: ticket.StatusTodo})
	if err := db.AddDependency(blocked.ID, blocker.ID, "test"); err != nil {
		t.Fatal(err)
	}

	var got readinessResponse
	if code := get(t, srv, "/api/v1/readiness", &got); code != http.StatusOK {
		t.Fatalf("readiness status = %d, want 200", code)
	}
	if got.Next == nil || got.Next.Ticket.ID != high.ID || got.Next.Brief.Node.ID != high.ID {
		t.Fatalf("next = %+v, want %s with its brief", got.Next, high.ID)
	}
	if len(got.Next.Brief.Acceptance) != 1 || got.Next.Brief.Acceptance[0] != "ship the ready surface" {
		t.Fatalf("next brief acceptance = %v", got.Next.Brief.Acceptance)
	}
	if len(got.Ready) != 2 || got.Ready[0].ID != high.ID || got.Ready[1].ID != low.ID {
		t.Fatalf("ready order = %v, want [%s %s]", ticketIDs(got.Ready), high.ID, low.ID)
	}
	if len(got.Blocked) != 1 || got.Blocked[0].ID != blocked.ID {
		t.Fatalf("blocked = %+v, want %s", got.Blocked, blocked.ID)
	}
	if len(got.Blocked[0].BlockedBy) != 1 || got.Blocked[0].BlockedBy[0].ID != blocker.ID || got.Blocked[0].BlockedBy[0].Status != "in_progress" {
		t.Fatalf("blocked_by = %+v, want %s (in_progress)", got.Blocked[0].BlockedBy, blocker.ID)
	}
}

func TestReadinessUsesEmptyCollectionsWhenNothingReady(t *testing.T) {
	srv, _ := newTestServer(t)
	var got readinessResponse
	if code := get(t, srv, "/api/v1/readiness", &got); code != http.StatusOK {
		t.Fatalf("readiness status = %d, want 200", code)
	}
	if got.Next != nil || got.Ready == nil || got.Blocked == nil || len(got.Ready) != 0 || len(got.Blocked) != 0 {
		t.Fatalf("empty readiness = %+v", got)
	}
}

func priority(value float64) *float64 { return &value }

func ticketIDs(nodes []*ticket.Ticket) []string {
	ids := make([]string, len(nodes))
	for i, node := range nodes {
		ids[i] = node.ID
	}
	return ids
}
