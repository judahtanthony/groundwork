package server

import (
	"net/http"
	"strings"
	"testing"

	"groundwork/internal/completion"
	"groundwork/internal/store/sqlite"
	"groundwork/internal/ticket"
)

// The review screen renders a feature-level bundle: recommendation plus each
// child's summary and validation (ADR 0057/T-1086).
func TestReviewPageRendersBundle(t *testing.T) {
	srv, db := newTestServer(t)
	root := mustCreate(t, db, &ticket.Ticket{Title: "feature", NodeType: ticket.NodeComposite, Status: ticket.StatusTodo, WorkType: "technical_design"})
	child := mustCreate(t, db, &ticket.Ticket{ParentID: root.ID, Title: "child A", NodeType: ticket.NodeLeaf, Status: ticket.StatusDone, WorkType: "technical_implementation"})
	if err := db.UpsertCompletionSummary(&completion.Summary{NodeID: child.ID, Outcome: "implemented child A"}); err != nil {
		t.Fatal(err)
	}
	if _, err := db.RecordValidation(sqlite.ValidationResult{TicketID: child.ID, Name: "go", Status: sqlite.ValidationPass}); err != nil {
		t.Fatal(err)
	}

	rr := getHTML(t, srv, "/review/"+root.ID)
	if rr.Code != http.StatusOK {
		t.Fatalf("GET /review = %d, want 200\n%s", rr.Code, rr.Body.String())
	}
	body := rr.Body.String()
	for _, want := range []string{root.ID, "Review", "land", child.ID, "implemented child A", "validation go: pass"} {
		if !strings.Contains(body, want) {
			t.Errorf("review page missing %q", want)
		}
	}
}
