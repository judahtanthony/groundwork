package server

import (
	"testing"

	"groundwork/internal/approval"
	"groundwork/internal/decision"
	"groundwork/internal/policy"
	"groundwork/internal/risk"
	"groundwork/internal/ticket"
)

// TestPendingApprovalEmitsDurableRecord proves a human-gated pending approval
// emits a paired durable approval_requested record (ADR 0051/0053), and that a
// terminal decision resolves it so it is not reprojected on rebuild (T-1059).
func TestPendingApprovalEmitsAndResolvesDurableRecord(t *testing.T) {
	svc, db := requestService(t)
	tk := &ticket.Ticket{Title: "risky", Status: ticket.StatusInProgress, WorkType: "technical_implementation"}
	if err := db.CreateTicket(tk, "tester"); err != nil {
		t.Fatal(err)
	}

	// A non-docs execute is not auto-approved by the docs policy: it is pending.
	a, err := svc.Request(RequestParams{
		TicketID: tk.ID, Type: approval.TypeExecute, Summary: "run risky step",
		Action: policy.Action{Type: "execute", ChangeType: "code", Scope: risk.Scope{Files: []string{"main.go"}}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if a.Status != string(approval.StatusPending) {
		t.Fatalf("status = %s, want pending", a.Status)
	}

	// A durable approval_requested record exists, correlated to the approval id.
	recs, err := db.ListDecisions(tk.ID)
	if err != nil || len(recs) != 1 {
		t.Fatalf("decisions = %+v err=%v", recs, err)
	}
	if recs[0].EventType != decision.EventApprovalRequested || recs[0].ID != a.ID || recs[0].Status != decision.StatusPending {
		t.Fatalf("durable record malformed: %+v", recs[0])
	}

	// It would reproject on rebuild while pending.
	pending, _ := db.ListPendingDecisions()
	if len(pending) != 1 {
		t.Fatalf("pending decisions = %d, want 1", len(pending))
	}

	// A terminal decision resolves the durable request.
	if _, err := svc.Decide(a.ID, approval.StatusApproved, "human.owner", "ok"); err != nil {
		t.Fatal(err)
	}
	pending, _ = db.ListPendingDecisions()
	if len(pending) != 0 {
		t.Fatalf("pending decisions after decide = %d, want 0 (resolved)", len(pending))
	}

	// And so it does NOT recreate a phantom approval on rebuild.
	rep, err := db.RebuildDurableQueues()
	if err != nil {
		t.Fatal(err)
	}
	if rep.ApprovalsRecreated != 0 {
		t.Fatalf("recreated %d phantom approvals, want 0", rep.ApprovalsRecreated)
	}
}
