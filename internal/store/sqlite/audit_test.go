package sqlite

import (
	"testing"

	"groundwork/internal/ticket"
)

func TestRecentAuditEventsNewestFirst(t *testing.T) {
	db := openTestDB(t)
	// CreateTicket appends a ticket.created audit event per node.
	for _, title := range []string{"a", "b", "c"} {
		if err := db.CreateTicket(&ticket.Ticket{Title: title, Status: ticket.StatusTodo}, "t"); err != nil {
			t.Fatal(err)
		}
	}

	ev, err := db.RecentAuditEvents(2)
	if err != nil {
		t.Fatal(err)
	}
	if len(ev) != 2 {
		t.Fatalf("got %d events, want 2 (limit honored)", len(ev))
	}
	if ev[0].ID < ev[1].ID {
		t.Errorf("events not newest-first: id %d before %d", ev[0].ID, ev[1].ID)
	}
	if ev[0].Type != "ticket.created" {
		t.Errorf("newest event type = %q, want ticket.created", ev[0].Type)
	}

	// A zero/negative limit falls back to a sane default rather than returning none.
	all, err := db.RecentAuditEvents(0)
	if err != nil {
		t.Fatal(err)
	}
	if len(all) != 3 {
		t.Errorf("default-limit events = %d, want 3", len(all))
	}
}
