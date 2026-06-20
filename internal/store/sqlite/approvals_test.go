package sqlite

import (
	"strings"
	"testing"

	"groundwork/internal/approval"
	"groundwork/internal/ticket"
)

func TestApprovalCreateDecide(t *testing.T) {
	db := openTestDB(t)
	tk := todoTicket(t, db, "node")

	a, err := db.CreateApproval(CreateApprovalParams{
		TicketID: tk.ID, Type: approval.TypeLandToMain, RiskClass: "medium",
		Summary: "land it", RequestedByActor: "ai.codex.default",
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if a.Status != string(approval.StatusPending) {
		t.Fatalf("status = %s, want pending", a.Status)
	}

	decided, err := db.DecideApproval(a.ID, approval.StatusApproved, "human.owner", "ok")
	if err != nil {
		t.Fatalf("decide: %v", err)
	}
	if decided.Status != string(approval.StatusApproved) || decided.DecidedByActor != "human.owner" {
		t.Fatalf("decided = %+v", decided)
	}
	// Deciding again is rejected.
	if _, err := db.DecideApproval(a.ID, approval.StatusRejected, "human.owner", ""); err == nil {
		t.Error("expected error deciding an already-decided approval")
	}
}

func TestApprovalListFilter(t *testing.T) {
	db := openTestDB(t)
	tk := todoTicket(t, db, "node")
	a1, _ := db.CreateApproval(CreateApprovalParams{TicketID: tk.ID, Type: approval.TypeExecute, RiskClass: "low", Summary: "a1", RequestedByActor: "x"})
	_, _ = db.CreateApproval(CreateApprovalParams{TicketID: tk.ID, Type: approval.TypeExecute, RiskClass: "low", Summary: "a2", RequestedByActor: "x"})
	_, _ = db.DecideApproval(a1.ID, approval.StatusApproved, "human.owner", "")

	pending, _ := db.ListApprovals(string(approval.StatusPending))
	if len(pending) != 1 {
		t.Errorf("pending = %d, want 1", len(pending))
	}
	all, _ := db.ListApprovals("")
	if len(all) != 2 {
		t.Errorf("all = %d, want 2", len(all))
	}
}

func compositeParent(t *testing.T, db *DB, title string) *ticket.Ticket {
	t.Helper()
	tk := &ticket.Ticket{Title: title, Status: ticket.StatusInProgress}
	if err := db.CreateTicket(tk, "tester"); err != nil {
		t.Fatal(err)
	}
	if err := db.TriageTicket(tk.ID, ticket.NodeComposite, "tester"); err != nil {
		t.Fatal(err)
	}
	return tk
}

func TestDecomposeProposalAndAccept(t *testing.T) {
	db := openTestDB(t)
	parent := compositeParent(t, db, "epic")

	appr, childIDs, err := db.DecomposeProposal(parent.ID, `{"schema":"contract/v1"}`,
		[]ChildSpec{{Title: "child a", WorkType: "technical_implementation"}, {Title: "child b"}}, "ai.codex.default")
	if err != nil {
		t.Fatalf("decompose: %v", err)
	}
	if len(childIDs) != 2 || appr.Type != string(approval.TypeDecompose) {
		t.Fatalf("appr=%+v childIDs=%v", appr, childIDs)
	}
	// Children start in backlog; parent moves to review.
	for _, cid := range childIDs {
		c, _ := db.GetTicket(cid)
		if c.Status != ticket.StatusBacklog {
			t.Errorf("child %s status = %s, want backlog", cid, c.Status)
		}
	}
	if p, _ := db.GetTicket(parent.ID); p.Status != ticket.StatusReview {
		t.Errorf("parent status = %s, want review", p.Status)
	}

	// Accept: children become todo, parent contract written, parent approved.
	if _, err := db.AcceptDecompose(appr.ID, "human.owner", "looks good"); err != nil {
		t.Fatalf("accept: %v", err)
	}
	for _, cid := range childIDs {
		c, _ := db.GetTicket(cid)
		if c.Status != ticket.StatusTodo {
			t.Errorf("child %s status = %s, want todo", cid, c.Status)
		}
	}
	p, _ := db.GetTicket(parent.ID)
	if p.Status != ticket.StatusApproved {
		t.Errorf("parent status = %s, want approved", p.Status)
	}
	if !strings.Contains(p.Contract, "contract/v1") {
		t.Errorf("parent contract not promoted: %q", p.Contract)
	}
}

func TestDecomposeReject(t *testing.T) {
	db := openTestDB(t)
	parent := compositeParent(t, db, "epic")
	appr, childIDs, _ := db.DecomposeProposal(parent.ID, "{}", []ChildSpec{{Title: "c"}}, "ai.codex.default")

	if _, err := db.RejectDecompose(appr.ID, "human.owner", "needs work"); err != nil {
		t.Fatalf("reject: %v", err)
	}
	if p, _ := db.GetTicket(parent.ID); p.Status != ticket.StatusRework {
		t.Errorf("parent status = %s, want rework", p.Status)
	}
	// Children remain in backlog (non-dispatchable).
	if c, _ := db.GetTicket(childIDs[0]); c.Status != ticket.StatusBacklog {
		t.Errorf("child status = %s, want backlog", c.Status)
	}
}

func TestEscalateAndReplan(t *testing.T) {
	db := openTestDB(t)
	tk := todoTicket(t, db, "node")

	appr, err := db.Escalate(tk.ID, "requirements changed", "ai.codex.default")
	if err != nil {
		t.Fatalf("escalate: %v", err)
	}
	if appr.Type != string(approval.TypeReplan) {
		t.Fatalf("appr type = %s, want replan", appr.Type)
	}
	if got, _ := db.GetTicket(tk.ID); got.Status != ticket.StatusBlocked {
		t.Errorf("ticket status = %s, want blocked", got.Status)
	}

	// Accept re-plan requeues the node.
	if _, err := db.AcceptReplan(appr.ID, "human.owner", "re-pointed"); err != nil {
		t.Fatalf("accept replan: %v", err)
	}
	if got, _ := db.GetTicket(tk.ID); got.Status != ticket.StatusTodo {
		t.Errorf("ticket status = %s, want todo after replan", got.Status)
	}
}
