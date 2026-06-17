package sqlite

import (
	"testing"

	"groundwork/internal/ticket"
)

func TestCreateAllocatesSequentialIDs(t *testing.T) {
	db := openTestDB(t)
	var ids []string
	for i := 0; i < 3; i++ {
		tk := &ticket.Ticket{Title: "node"}
		if err := db.CreateTicket(tk, "human"); err != nil {
			t.Fatalf("CreateTicket: %v", err)
		}
		ids = append(ids, tk.ID)
	}
	want := []string{"T-0001", "T-0002", "T-0003"}
	for i := range want {
		if ids[i] != want[i] {
			t.Errorf("id[%d] = %q, want %q", i, ids[i], want[i])
		}
	}
}

func TestCreateAppliesDefaultsAndRoundTrips(t *testing.T) {
	db := openTestDB(t)
	p := 2
	in := &ticket.Ticket{
		Title:      "Build the thing",
		Labels:     []string{"store", "sqlite"},
		Acceptance: []string{"migrations apply", "re-run safe"},
		Priority:   &p,
	}
	if err := db.CreateTicket(in, "human"); err != nil {
		t.Fatalf("CreateTicket: %v", err)
	}

	got, err := db.GetTicket(in.ID)
	if err != nil {
		t.Fatalf("GetTicket: %v", err)
	}
	if got.Kind != "ticket" {
		t.Errorf("kind default = %q, want ticket", got.Kind)
	}
	if got.Status != ticket.StatusBacklog {
		t.Errorf("status default = %q, want backlog", got.Status)
	}
	if got.Contract != "{}" {
		t.Errorf("contract default = %q, want {}", got.Contract)
	}
	if len(got.Labels) != 2 || got.Labels[0] != "store" {
		t.Errorf("labels = %v", got.Labels)
	}
	if len(got.Acceptance) != 2 || got.Acceptance[1] != "re-run safe" {
		t.Errorf("acceptance = %v", got.Acceptance)
	}
	if got.Priority == nil || *got.Priority != 2 {
		t.Errorf("priority = %v, want 2", got.Priority)
	}
	if got.CreatedAt == "" || got.UpdatedAt == "" {
		t.Error("timestamps not set")
	}
}

func TestEveryMutationAppendsAudit(t *testing.T) {
	db := openTestDB(t)
	tk := &ticket.Ticket{Title: "node"}
	if err := db.CreateTicket(tk, "human"); err != nil {
		t.Fatal(err)
	}
	tk.Title = "renamed"
	if err := db.UpdateTicket(tk, "human"); err != nil {
		t.Fatal(err)
	}

	events, err := db.AuditEventsFor("ticket", tk.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 2 {
		t.Fatalf("audit events = %d, want 2 (create+update)", len(events))
	}
	if events[0].Type != "ticket.created" || events[1].Type != "ticket.updated" {
		t.Errorf("audit types = %q, %q", events[0].Type, events[1].Type)
	}
}

func TestSeedTicketSeqNeverLowers(t *testing.T) {
	db := openTestDB(t)
	if err := db.SeedTicketSeq(50); err != nil {
		t.Fatal(err)
	}
	if err := db.SeedTicketSeq(10); err != nil { // lower: ignored
		t.Fatal(err)
	}
	tk := &ticket.Ticket{Title: "after seed"}
	if err := db.CreateTicket(tk, "human"); err != nil {
		t.Fatal(err)
	}
	if tk.ID != "T-0051" {
		t.Errorf("id after seed = %q, want T-0051", tk.ID)
	}
}

func TestParentChildPersists(t *testing.T) {
	db := openTestDB(t)
	parent := &ticket.Ticket{Title: "parent", NodeType: ticket.NodeComposite}
	if err := db.CreateTicket(parent, "human"); err != nil {
		t.Fatal(err)
	}
	child := &ticket.Ticket{Title: "child", ParentID: parent.ID}
	if err := db.CreateTicket(child, "human"); err != nil {
		t.Fatal(err)
	}
	got, err := db.GetTicket(child.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.ParentID != parent.ID {
		t.Errorf("parent_id = %q, want %q", got.ParentID, parent.ID)
	}
}
