package sqlite

import (
	"testing"

	"groundwork/internal/completion"
	"groundwork/internal/ticket"
)

func TestCompletionSummaryMirror(t *testing.T) {
	db := openTestDB(t)
	tk := &ticket.Ticket{Title: "child", NodeType: ticket.NodeLeaf, Status: ticket.StatusInProgress, WorkType: "technical_implementation"}
	if err := db.CreateTicket(tk, "tester"); err != nil {
		t.Fatal(err)
	}
	if got, err := db.GetCompletionSummary(tk.ID); err != nil || got != nil {
		t.Fatalf("before: got=%v err=%v, want nil", got, err)
	}
	want := &completion.Summary{NodeID: tk.ID, Outcome: "done", Changed: []string{"a.go"}, Risks: []string{"r1"}}
	if err := db.UpsertCompletionSummary(want); err != nil {
		t.Fatal(err)
	}
	got, err := db.GetCompletionSummary(tk.ID)
	if err != nil || got == nil {
		t.Fatalf("after: got=%v err=%v", got, err)
	}
	if got.Outcome != "done" || len(got.Changed) != 1 || len(got.Risks) != 1 {
		t.Errorf("mirror mismatch: %+v", got)
	}
}
