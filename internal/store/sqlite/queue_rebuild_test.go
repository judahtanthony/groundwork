package sqlite

import (
	"testing"

	"groundwork/internal/approval"
	"groundwork/internal/decision"
	"groundwork/internal/ticket"
)

func seedTicketStatus(t *testing.T, db *DB, title string, status ticket.Status) string {
	t.Helper()
	tk := &ticket.Ticket{Title: title, NodeType: ticket.NodeLeaf, Status: status, WorkType: "technical_implementation"}
	if err := db.CreateTicket(tk, "tester"); err != nil {
		t.Fatal(err)
	}
	return tk.ID
}

// importApprovalRequest mimics a sidecar-imported pending approval_requested
// record (the durable form that survives a store purge).
func importApprovalRequest(t *testing.T, db *DB, ticketID, reqType, summary string) {
	t.Helper()
	rev := true
	rec := decision.Record{
		Sequence: 1, ID: "D-" + ticketID, EventType: decision.EventApprovalRequested,
		TicketID: ticketID, RequestType: reqType, Status: decision.StatusPending,
		RequestedBy: "ai.codex.default", RequestedActor: "human.owner", Statement: summary,
		PolicyInputs: &decision.PolicyInputs{Action: reqType, RiskClass: "medium", Reversible: &rev},
	}
	if err := db.ImportDecision(rec); err != nil {
		t.Fatal(err)
	}
}

// TestRebuildDurableQueuesRecreatesApprovals covers the cold-rebuild projection
// for decompose, replan, and land_to_main pending approval requests (T-1054).
func TestRebuildDurableQueuesRecreatesApprovals(t *testing.T) {
	db := openTestDB(t)
	dec := seedTicketStatus(t, db, "decompose me", ticket.StatusReview)
	rep := seedTicketStatus(t, db, "replan me", ticket.StatusReview)
	land := seedTicketStatus(t, db, "land me", ticket.StatusReview)
	importApprovalRequest(t, db, dec, "decompose", "Accept children?")
	importApprovalRequest(t, db, rep, "replan", "Accept re-plan?")
	importApprovalRequest(t, db, land, "land_to_main", "Land to main?")

	// Approvals table is empty (purged); rebuild recreates the three rows.
	report, err := db.RebuildDurableQueues()
	if err != nil {
		t.Fatal(err)
	}
	if report.ApprovalsRecreated != 3 {
		t.Fatalf("recreated %d approvals, want 3", report.ApprovalsRecreated)
	}
	pending, err := db.ListApprovals(string(approval.StatusPending))
	if err != nil {
		t.Fatal(err)
	}
	if len(pending) != 3 {
		t.Fatalf("pending approvals %d, want 3", len(pending))
	}
	for _, a := range pending {
		if a.ID == "" || a.RiskClass != "medium" {
			t.Errorf("recreated approval malformed: %+v", a)
		}
	}

	// Idempotent: a second run recreates nothing (the rows now exist).
	report2, err := db.RebuildDurableQueues()
	if err != nil {
		t.Fatal(err)
	}
	if report2.ApprovalsRecreated != 0 {
		t.Fatalf("second run recreated %d, want 0", report2.ApprovalsRecreated)
	}
}

// TestRebuildDurableQueuesInputRequiredKeepsBlocked covers a blocked ticket whose
// durable explainer is an input_requested record: it needs no approval row and
// must NOT be flagged recovery_needed (T-1054).
func TestRebuildDurableQueuesInputRequiredExplained(t *testing.T) {
	db := openTestDB(t)
	id := seedTicketStatus(t, db, "blocked on input", ticket.StatusBlocked)
	if err := db.ImportDecision(decision.Record{
		Sequence: 1, ID: "D-1", EventType: decision.EventInputRequested, TicketID: id,
		Status: decision.StatusPending, Statement: "Which timeout key?",
	}); err != nil {
		t.Fatal(err)
	}
	report, err := db.RebuildDurableQueues()
	if err != nil {
		t.Fatal(err)
	}
	if report.RecoveryNeeded != 0 {
		t.Fatalf("recovery_needed %d, want 0 (input request explains the block)", report.RecoveryNeeded)
	}
}

// TestRebuildDurableQueuesSurfacesRecoveryNeeded covers a blocked ticket with no
// durable explainer and a review ticket with no pending request: both get a
// recovery_needed record instead of silently stranding (T-1054).
func TestRebuildDurableQueuesSurfacesRecoveryNeeded(t *testing.T) {
	db := openTestDB(t)
	blocked := seedTicketStatus(t, db, "stranded blocked", ticket.StatusBlocked)
	review := seedTicketStatus(t, db, "stranded review", ticket.StatusReview)
	_ = seedTicketStatus(t, db, "healthy todo", ticket.StatusTodo) // must not be flagged

	report, err := db.RebuildDurableQueues()
	if err != nil {
		t.Fatal(err)
	}
	if report.RecoveryNeeded != 2 {
		t.Fatalf("recovery_needed %d, want 2", report.RecoveryNeeded)
	}
	for _, id := range []string{blocked, review} {
		recs, err := db.ListDecisions(id)
		if err != nil {
			t.Fatal(err)
		}
		if len(recs) != 1 || recs[0].EventType != decision.EventRecoveryNeeded {
			t.Fatalf("%s: expected one recovery_needed record, got %+v", id, recs)
		}
	}
	// Idempotent: the recovery_needed record is now itself the explainer.
	report2, err := db.RebuildDurableQueues()
	if err != nil {
		t.Fatal(err)
	}
	if report2.RecoveryNeeded != 0 {
		t.Fatalf("second run flagged %d, want 0", report2.RecoveryNeeded)
	}
}
