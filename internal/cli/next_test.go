package cli

import (
	"testing"

	"groundwork/internal/ticket"
)

func TestNextNodePicksValueOrderedTop(t *testing.T) {
	db := cliTestDB(t)

	mk := func(prio float64) *ticket.Ticket {
		tk := &ticket.Ticket{Title: "n", Status: ticket.StatusTodo, WorkType: "technical_implementation"}
		if prio > 0 {
			p := prio
			tk.Priority = &p
		}
		if err := db.CreateTicket(tk, "t"); err != nil {
			t.Fatalf("create: %v", err)
		}
		return tk
	}
	mk(0.2)
	hi := mk(0.8)
	mk(0.5)

	top, err := nextNode(db)
	if err != nil {
		t.Fatal(err)
	}
	if top == nil || top.ID != hi.ID {
		t.Fatalf("next = %v, want highest-priority %s", top, hi.ID)
	}
}

func TestNextNodeEmptyWhenNothingReady(t *testing.T) {
	db := cliTestDB(t)

	// A todo node blocked by an unmet dependency is not eligible.
	dep := &ticket.Ticket{Title: "dep", Status: ticket.StatusTodo, WorkType: "technical_implementation"}
	if err := db.CreateTicket(dep, "t"); err != nil {
		t.Fatal(err)
	}
	blocked := &ticket.Ticket{Title: "blocked", Status: ticket.StatusTodo, WorkType: "technical_implementation"}
	if err := db.CreateTicket(blocked, "t"); err != nil {
		t.Fatal(err)
	}
	if err := db.AddDependency(blocked.ID, dep.ID, "t"); err != nil {
		t.Fatal(err)
	}
	// dep itself is eligible; move it out of todo so nothing is ready.
	if err := db.TransitionTicket(dep.ID, ticket.StatusInProgress, "t"); err != nil {
		t.Fatal(err)
	}

	top, err := nextNode(db)
	if err != nil {
		t.Fatal(err)
	}
	if top != nil {
		t.Fatalf("next = %s, want nil (nothing ready)", top.ID)
	}
}
