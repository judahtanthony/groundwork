package sqlite

import (
	"testing"

	"groundwork/internal/ticket"
)

func TestEligibilityRequiresTodo(t *testing.T) {
	db := openTestDB(t)
	tk := &ticket.Ticket{Title: "node", Status: ticket.StatusBacklog}
	if err := db.CreateTicket(tk, "human"); err != nil {
		t.Fatal(err)
	}
	ok, err := db.IsEligible(tk.ID)
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Error("backlog node should not be eligible")
	}
}

func TestEligibilityRecomputesAsDepsComplete(t *testing.T) {
	db := openTestDB(t)
	dep := &ticket.Ticket{Title: "dep", Status: ticket.StatusInProgress}
	node := &ticket.Ticket{Title: "node", Status: ticket.StatusTodo}
	if err := db.CreateTicket(dep, "human"); err != nil {
		t.Fatal(err)
	}
	if err := db.CreateTicket(node, "human"); err != nil {
		t.Fatal(err)
	}
	if err := db.AddDependency(node.ID, dep.ID, "human"); err != nil {
		t.Fatal(err)
	}

	// Dependency not done: node ineligible.
	if ok, _ := db.IsEligible(node.ID); ok {
		t.Fatal("node should be ineligible while dependency is in progress")
	}

	// Complete the dependency.
	if err := db.TransitionTicket(dep.ID, ticket.StatusDone, "human"); err != nil {
		t.Fatal(err)
	}
	if ok, _ := db.IsEligible(node.ID); !ok {
		t.Fatal("node should become eligible once dependency is done")
	}

	eligible, _ := db.ListEligible()
	if len(eligible) != 1 || eligible[0].ID != node.ID {
		t.Fatalf("ListEligible = %v, want [%s]", ids(eligible), node.ID)
	}
}
