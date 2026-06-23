package cli

import (
	"testing"

	"groundwork/internal/ticket"
)

func TestReadinessCounts(t *testing.T) {
	db := cliTestDB(t)
	// ready node, a done dep, and a node blocked by an unmet (todo) dep.
	_, _, _, _ = readyBlockedFixture(t, db) // done, ready, blocked(by pending), pending dep

	all, err := db.ListTickets()
	if err != nil {
		t.Fatal(err)
	}
	eligible, blocked, pending, err := readinessCounts(db, all)
	if err != nil {
		t.Fatal(err)
	}
	// Fixture todo nodes: "ready" (deps satisfied) and "pending dep" (no deps) are
	// eligible; "blocked node" is blocked by the pending dep.
	if eligible != 2 {
		t.Errorf("eligible = %d, want 2", eligible)
	}
	if blocked != 1 {
		t.Errorf("blocked = %d, want 1", blocked)
	}
	if pending != 0 {
		t.Errorf("pending approvals = %d, want 0", pending)
	}
}

func TestReadinessCountsIgnoresNonTodo(t *testing.T) {
	db := cliTestDB(t)
	mk := func(s ticket.Status) {
		if err := db.CreateTicket(&ticket.Ticket{Title: "n", Status: s, WorkType: "technical_implementation"}, "t"); err != nil {
			t.Fatal(err)
		}
	}
	mk(ticket.StatusBacklog)
	mk(ticket.StatusInProgress)
	mk(ticket.StatusDone)
	mk(ticket.StatusTodo) // the only eligible one

	all, _ := db.ListTickets()
	eligible, blocked, _, err := readinessCounts(db, all)
	if err != nil {
		t.Fatal(err)
	}
	if eligible != 1 || blocked != 0 {
		t.Errorf("eligible=%d blocked=%d, want 1/0", eligible, blocked)
	}
}
