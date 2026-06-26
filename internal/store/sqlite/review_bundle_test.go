package sqlite

import (
	"testing"

	"groundwork/internal/approval"
	"groundwork/internal/completion"
	"groundwork/internal/ticket"
)

func TestReviewBundle(t *testing.T) {
	db := openTestDB(t)
	root := &ticket.Ticket{Title: "feature", NodeType: ticket.NodeComposite, Status: ticket.StatusTodo, WorkType: "technical_design"}
	if err := db.CreateTicket(root, "tester"); err != nil {
		t.Fatal(err)
	}
	var leaves []*ticket.Ticket
	for i := 0; i < 2; i++ {
		c := &ticket.Ticket{ParentID: root.ID, Title: "child", NodeType: ticket.NodeLeaf, Status: ticket.StatusDone, WorkType: "technical_implementation"}
		if err := db.CreateTicket(c, "tester"); err != nil {
			t.Fatal(err)
		}
		if err := db.UpsertCompletionSummary(&completion.Summary{NodeID: c.ID, Outcome: "did it"}); err != nil {
			t.Fatal(err)
		}
		if _, err := db.RecordValidation(ValidationResult{TicketID: c.ID, Name: "go", Status: ValidationPass}); err != nil {
			t.Fatal(err)
		}
		leaves = append(leaves, c)
	}

	// Clean subtree -> land.
	b, err := db.ReviewBundle(root.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(b.Children) != 2 || b.Recommendation != "land" {
		t.Fatalf("clean bundle: children=%d rec=%s, want 2/land", len(b.Children), b.Recommendation)
	}
	if b.Children[0].Summary == nil || b.Children[0].Summary.Outcome != "did it" {
		t.Errorf("child summary not aggregated: %+v", b.Children[0])
	}

	// A failing validation -> rework.
	if _, err := db.RecordValidation(ValidationResult{TicketID: leaves[1].ID, Name: "lint", Status: ValidationFail}); err != nil {
		t.Fatal(err)
	}
	if b, _ := db.ReviewBundle(root.ID); b.Recommendation != "rework" {
		t.Errorf("with failure: rec=%s, want rework", b.Recommendation)
	}

	// A pending exception -> hold (outranks rework).
	if _, err := db.CreateApproval(CreateApprovalParams{
		TicketID: leaves[0].ID, Type: approval.TypeException, RiskClass: "medium",
		Summary: "exceeds envelope", Status: approval.StatusPending, RequestedByActor: "ai.coding.codex",
	}); err != nil {
		t.Fatal(err)
	}
	b, _ = db.ReviewBundle(root.ID)
	if b.Recommendation != "hold" || len(b.UnresolvedExceptions) != 1 {
		t.Errorf("with exception: rec=%s unresolved=%d, want hold/1", b.Recommendation, len(b.UnresolvedExceptions))
	}
}
