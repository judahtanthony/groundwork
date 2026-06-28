package sqlite

import (
	"testing"

	"groundwork/internal/decision"
	"groundwork/internal/ticket"
)

func seedTicket(t *testing.T, db *DB, title string) string {
	t.Helper()
	tk := &ticket.Ticket{Title: title, NodeType: ticket.NodeLeaf, Status: ticket.StatusInProgress, WorkType: "technical_implementation"}
	if err := db.CreateTicket(tk, "tester"); err != nil {
		t.Fatal(err)
	}
	return tk.ID
}

func TestAppendDecisionAssignsSequence(t *testing.T) {
	db := openTestDB(t)
	id := seedTicket(t, db, "child")

	r1, err := db.AppendDecision(decision.Record{
		EventType: decision.EventInputRequested, TicketID: id, Status: decision.StatusPending,
		Statement: "first", RequestedAt: "2026-06-24T15:00:00Z",
	})
	if err != nil {
		t.Fatal(err)
	}
	if r1.Sequence != 1 {
		t.Fatalf("first seq = %d, want 1", r1.Sequence)
	}
	r2, err := db.AppendDecision(decision.Record{
		EventType: decision.EventApprovalRequested, TicketID: id, Status: decision.StatusPending,
		Statement: "second", RequestedAt: "2026-06-24T15:01:00Z",
	})
	if err != nil {
		t.Fatal(err)
	}
	if r2.Sequence != 2 {
		t.Fatalf("second seq = %d, want 2", r2.Sequence)
	}

	got, err := db.ListDecisions(id)
	if err != nil || len(got) != 2 {
		t.Fatalf("list: n=%d err=%v", len(got), err)
	}
	if got[0].Statement != "first" || got[1].Statement != "second" {
		t.Fatalf("order wrong: %+v", got)
	}
}

func TestListPendingDecisions(t *testing.T) {
	db := openTestDB(t)
	a := seedTicket(t, db, "a")
	b := seedTicket(t, db, "b")
	if _, err := db.AppendDecision(decision.Record{EventType: decision.EventApprovalRequested, TicketID: a, Status: decision.StatusPending, Statement: "p1"}); err != nil {
		t.Fatal(err)
	}
	if _, err := db.AppendDecision(decision.Record{EventType: decision.EventApprovalDecided, TicketID: a, Status: decision.StatusAccepted, Statement: "decided"}); err != nil {
		t.Fatal(err)
	}
	if _, err := db.AppendDecision(decision.Record{EventType: decision.EventInputRequested, TicketID: b, Status: decision.StatusPending, Statement: "p2"}); err != nil {
		t.Fatal(err)
	}
	pending, err := db.ListPendingDecisions()
	if err != nil {
		t.Fatal(err)
	}
	if len(pending) != 2 {
		t.Fatalf("pending n=%d, want 2", len(pending))
	}
}

func TestImportDecisionPreservesSequence(t *testing.T) {
	db := openTestDB(t)
	id := seedTicket(t, db, "child")
	rec := decision.Record{
		ID: "D-0042", Sequence: 7, EventType: decision.EventRecoveryNeeded, TicketID: id,
		Status: decision.StatusPending, HandoffSummary: "lost context",
	}
	if err := db.ImportDecision(rec); err != nil {
		t.Fatal(err)
	}
	// Idempotent re-import (upsert on ticket_id, seq).
	if err := db.ImportDecision(rec); err != nil {
		t.Fatal(err)
	}
	got, err := db.ListDecisions(id)
	if err != nil || len(got) != 1 {
		t.Fatalf("list n=%d err=%v", len(got), err)
	}
	if got[0].Sequence != 7 || got[0].ID != "D-0042" {
		t.Fatalf("not preserved: %+v", got[0])
	}
	// A subsequent AppendDecision continues past the imported max sequence.
	r, err := db.AppendDecision(decision.Record{EventType: decision.EventInputRequested, TicketID: id, Status: decision.StatusPending, Statement: "next"})
	if err != nil {
		t.Fatal(err)
	}
	if r.Sequence != 8 {
		t.Fatalf("append after import seq = %d, want 8", r.Sequence)
	}
}

func TestImportDecisionRequiresSequence(t *testing.T) {
	db := openTestDB(t)
	id := seedTicket(t, db, "child")
	err := db.ImportDecision(decision.Record{EventType: decision.EventInputRequested, TicketID: id, Status: decision.StatusPending, Statement: "x"})
	if err == nil {
		t.Fatal("expected error importing record with seq 0")
	}
}
