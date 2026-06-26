package server

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"groundwork/internal/approval"
	"groundwork/internal/envelope"
	"groundwork/internal/ticket"
)

// Landing a root with an integration branch merges that branch into the default
// branch (--no-ff), deletes it, and closes the integration record (ADR 0058).
func TestRootLandMergesIntegrationBranchToMain(t *testing.T) {
	srv, db, root := newGitServer(t)
	def := srv.repo.DefaultBranch()
	if def == "" {
		t.Skip("no default branch in test repo")
	}

	parent := &ticket.Ticket{Title: "feature", NodeType: ticket.NodeComposite, Status: ticket.StatusTodo, WorkType: "technical_design"}
	if err := db.CreateTicket(parent, "tester"); err != nil {
		t.Fatal(err)
	}
	appr, err := srv.ProposeEnvelope(parent.ID, &envelope.Envelope{
		ApprovedActions: []string{envelope.ActionLandChildToParent}, AllowedRoles: []string{"coding"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := srv.recordDecision(appr.ID, approval.StatusApproved, ""); err != nil {
		t.Fatal(err)
	}
	ib, _ := db.GetIntegrationBranch(parent.ID)

	// A child lands its work to the integration branch.
	child := &ticket.Ticket{ParentID: parent.ID, Title: "child", NodeType: ticket.NodeLeaf, Status: ticket.StatusInProgress, WorkType: "technical_implementation"}
	if err := db.CreateTicket(child, "tester"); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "feature.go"), []byte("package x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	runGit(t, root, "add", "feature.go")
	if _, err := srv.LandToParent(child.ID); err != nil {
		t.Fatal(err)
	}

	// Land the root to main through the gate.
	for _, st := range []ticket.Status{ticket.StatusInProgress, ticket.StatusReview} {
		if err := db.TransitionTicket(parent.ID, st, "tester"); err != nil {
			t.Fatal(err)
		}
	}
	land, err := srv.approvals.RequestLanding(parent.ID, parent.WorkType)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := srv.recordDecision(land.ID, approval.StatusApproved, ""); err != nil {
		t.Fatal(err)
	}

	// On the default branch, integration branch gone, record closed, work merged.
	if cur, _ := srv.repo.CurrentBranch(); cur != def {
		t.Errorf("current branch = %q, want default %q after root landing", cur, def)
	}
	branches := runGit(t, root, "branch", "--list", ib.Branch)
	if strings.TrimSpace(branches) != "" {
		t.Errorf("integration branch %q still exists after landing", ib.Branch)
	}
	if rec, _ := db.GetIntegrationBranch(parent.ID); rec == nil || rec.Status != "landed" {
		t.Errorf("integration record = %+v, want status landed", rec)
	}
	if _, err := os.Stat(filepath.Join(root, "feature.go")); err != nil {
		t.Errorf("merged work missing on default branch: %v", err)
	}
}
