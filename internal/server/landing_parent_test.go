package server

import (
	"os"
	"path/filepath"
	"testing"

	"groundwork/internal/approval"
	"groundwork/internal/envelope"
	"groundwork/internal/ticket"
)

// A child lands to its root integration branch (not main): it becomes done and a
// commit advances the integration branch (ADR 0058).
func TestLandToParentCommitsToIntegrationBranch(t *testing.T) {
	srv, db, root := newGitServer(t)
	// Root + approved envelope → integration branch created and checked out.
	parent := &ticket.Ticket{Title: "feature", NodeType: ticket.NodeComposite, Status: ticket.StatusTodo, WorkType: "technical_design"}
	if err := db.CreateTicket(parent, "tester"); err != nil {
		t.Fatal(err)
	}
	appr, err := srv.ProposeEnvelope(parent.ID, &envelope.Envelope{
		ApprovedActions: []string{envelope.ActionExecuteChildren, envelope.ActionLandChildToParent},
		AllowedRoles:    []string{"coding"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := srv.recordDecision(appr.ID, approval.StatusApproved, ""); err != nil {
		t.Fatal(err)
	}
	ib, _ := db.GetIntegrationBranch(parent.ID)

	// A child leaf with staged work.
	child := &ticket.Ticket{ParentID: parent.ID, Title: "child", NodeType: ticket.NodeLeaf, Status: ticket.StatusInProgress, WorkType: "technical_implementation"}
	if err := db.CreateTicket(child, "tester"); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "feature.go"), []byte("package x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	runGit(t, root, "add", "feature.go")

	before := runGit(t, root, "rev-parse", "HEAD")
	got, err := srv.LandToParent(child.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Branch != ib.Branch {
		t.Errorf("landed to %q, want integration branch %q", got.Branch, ib.Branch)
	}
	if cur, _ := srv.repo.CurrentBranch(); cur != ib.Branch || cur == "main" || cur == "master" {
		t.Errorf("current branch = %q, want integration branch (not main)", cur)
	}
	if after := runGit(t, root, "rev-parse", "HEAD"); after == before {
		t.Error("HEAD did not advance; child work not committed to the integration branch")
	}
	if c, _ := db.GetTicket(child.ID); c.Status != ticket.StatusDone {
		t.Errorf("child status = %s, want done", c.Status)
	}
}

// land_to_parent requires an integration target (an approved root envelope).
func TestLandToParentWithoutTargetErrors(t *testing.T) {
	srv, db, _ := newGitServer(t)
	tk := &ticket.Ticket{Title: "orphan", NodeType: ticket.NodeLeaf, Status: ticket.StatusInProgress, WorkType: "technical_implementation"}
	if err := db.CreateTicket(tk, "tester"); err != nil {
		t.Fatal(err)
	}
	if _, err := srv.LandToParent(tk.ID); err == nil {
		t.Error("expected error landing without an integration target")
	}
}
