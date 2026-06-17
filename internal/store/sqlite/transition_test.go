package sqlite

import (
	"errors"
	"testing"

	"groundwork/internal/ticket"
)

func TestTransitionLegalAndAudited(t *testing.T) {
	db := openTestDB(t)
	tk := &ticket.Ticket{Title: "node", Status: ticket.StatusTodo}
	if err := db.CreateTicket(tk, "human"); err != nil {
		t.Fatal(err)
	}
	if err := db.TransitionTicket(tk.ID, ticket.StatusInProgress, "human"); err != nil {
		t.Fatalf("legal transition failed: %v", err)
	}
	got, _ := db.GetTicket(tk.ID)
	if got.Status != ticket.StatusInProgress {
		t.Errorf("status = %q, want in_progress", got.Status)
	}
	events, _ := db.AuditEventsFor("ticket", tk.ID)
	if events[len(events)-1].Type != "ticket.transitioned" {
		t.Errorf("last audit = %q, want ticket.transitioned", events[len(events)-1].Type)
	}
}

func TestTransitionIllegalRejected(t *testing.T) {
	db := openTestDB(t)
	tk := &ticket.Ticket{Title: "node", Status: ticket.StatusBacklog}
	if err := db.CreateTicket(tk, "human"); err != nil {
		t.Fatal(err)
	}
	// backlog -> done is not a legal manual transition.
	err := db.TransitionTicket(tk.ID, ticket.StatusDone, "human")
	if !errors.Is(err, ErrIllegalTransition) {
		t.Fatalf("want ErrIllegalTransition, got %v", err)
	}
}

func TestTriageSetsNodeType(t *testing.T) {
	db := openTestDB(t)
	tk := &ticket.Ticket{Title: "node"}
	if err := db.CreateTicket(tk, "human"); err != nil {
		t.Fatal(err)
	}
	if err := db.TriageTicket(tk.ID, ticket.NodeComposite, "human"); err != nil {
		t.Fatalf("triage: %v", err)
	}
	got, _ := db.GetTicket(tk.ID)
	if got.NodeType != ticket.NodeComposite {
		t.Errorf("node_type = %q, want composite", got.NodeType)
	}
}

func TestTriageRejectsBadType(t *testing.T) {
	db := openTestDB(t)
	tk := &ticket.Ticket{Title: "node"}
	if err := db.CreateTicket(tk, "human"); err != nil {
		t.Fatal(err)
	}
	if err := db.TriageTicket(tk.ID, ticket.NodeType("banana"), "human"); err == nil {
		t.Fatal("expected error for invalid node type")
	}
}
