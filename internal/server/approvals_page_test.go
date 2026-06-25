package server

import (
	"net/http"
	"strings"
	"testing"

	"groundwork/internal/approval"
	"groundwork/internal/store/sqlite"
	"groundwork/internal/ticket"
)

// The approvals inbox lists pending approvals grouped by risk, with the gate
// reason, requesting actor, and ticket context for each decision.
func TestApprovalsInboxRendersPending(t *testing.T) {
	srv, db := newTestServer(t)

	tk := &ticket.Ticket{Title: "Land me", Status: ticket.StatusInProgress, WorkType: "technical_implementation"}
	if err := db.CreateTicket(tk, "tester"); err != nil {
		t.Fatal(err)
	}
	if _, err := db.CreateApproval(sqlite.CreateApprovalParams{
		TicketID:         tk.ID,
		Type:             approval.TypeLandToMain,
		RiskClass:        "high",
		Summary:          "land node to main",
		Status:           approval.StatusPending,
		RequestedByActor: "ai.codex.default",
	}); err != nil {
		t.Fatal(err)
	}

	rr := getHTML(t, srv, "/approvals")
	if rr.Code != http.StatusOK {
		t.Fatalf("GET /approvals = %d, want 200\n%s", rr.Code, rr.Body.String())
	}
	body := rr.Body.String()
	for _, want := range []string{
		"high risk",                  // risk grouping
		"land_to_main",               // type
		"land node to main",          // summary
		"human-gated (land_to_main)", // derived gate reason
		"ai.codex.default",           // requesting actor
		tk.ID, "Land me",             // ticket context
	} {
		if !strings.Contains(body, want) {
			t.Errorf("approvals inbox missing %q", want)
		}
	}
}

// With nothing pending, the inbox shows its empty state rather than risk groups.
func TestApprovalsInboxEmptyState(t *testing.T) {
	srv, _ := newTestServer(t)
	body := getHTML(t, srv, "/approvals").Body.String()
	if !strings.Contains(body, "No pending approvals") {
		t.Errorf("approvals inbox missing empty state")
	}
}
