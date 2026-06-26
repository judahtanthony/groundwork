package server

import (
	"testing"

	"groundwork/internal/approval"
	"groundwork/internal/envelope"
	"groundwork/internal/ticket"
)

func TestSlugify(t *testing.T) {
	cases := map[string]string{
		"Operator UI":            "operator-ui",
		"  Land_to_parent path!": "land-to-parent-path",
		"":                       "",
	}
	for in, want := range cases {
		if got := slugify(in); got != want {
			t.Errorf("slugify(%q) = %q, want %q", in, got, want)
		}
	}
}

// Approving a root envelope on the default branch starts and records a
// gw/root/<id>-<slug> integration target (ADR 0058).
func TestEnsureIntegrationBranchOnEnvelopeApproval(t *testing.T) {
	srv, db, _ := newGitServer(t)
	parent := &ticket.Ticket{Title: "Operator UI", NodeType: ticket.NodeComposite, Status: ticket.StatusTodo, WorkType: "technical_design"}
	if err := db.CreateTicket(parent, "tester"); err != nil {
		t.Fatal(err)
	}
	appr, err := srv.ProposeEnvelope(parent.ID, &envelope.Envelope{
		ApprovedActions: []string{envelope.ActionExecuteChildren}, AllowedRoles: []string{"coding"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := srv.recordDecision(appr.ID, approval.StatusApproved, ""); err != nil {
		t.Fatal(err)
	}

	ib, err := db.GetIntegrationBranch(parent.ID)
	if err != nil || ib == nil {
		t.Fatalf("integration branch: got=%v err=%v", ib, err)
	}
	want := "gw/root/" + parent.ID + "-operator-ui"
	if ib.Branch != want || ib.Status != "open" || ib.BaseCommit == "" {
		t.Errorf("integration branch = %+v, want branch=%s open with base", ib, want)
	}
	if cur, _ := srv.repo.CurrentBranch(); cur != want {
		t.Errorf("current branch = %q, want %q (created and checked out)", cur, want)
	}
}
