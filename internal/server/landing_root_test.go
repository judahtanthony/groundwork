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

// A conflicting root land_to_main merge must abort cleanly: the work tree is
// restored, the integration branch is preserved, and the record stays open so the
// human can resolve and retry — a failed merge never leaves a mid-conflict tree
// or destroys the branch (H3/ADR 0058).
func TestRootLandAbortsOnMergeConflict(t *testing.T) {
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

	// Commit a change on the integration branch (checked out after approval).
	if err := os.WriteFile(filepath.Join(root, "feature.go"), []byte("package x // integration\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	runGit(t, root, "add", "feature.go")
	runGit(t, root, "commit", "-m", "integration change")

	// Commit a conflicting add of the same file on the default branch.
	runGit(t, root, "checkout", def)
	if err := os.WriteFile(filepath.Join(root, "feature.go"), []byte("package x // main\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	runGit(t, root, "add", "feature.go")
	runGit(t, root, "commit", "-m", "main change")

	if err := srv.mergeRootToMain(parent.ID); err == nil {
		t.Fatal("expected a merge-conflict error; got nil")
	}
	// The abort must clear the in-progress merge: no MERGE_HEAD and no unmerged
	// (conflicted) index entries. (The work tree shows .groundwork/ churn, so a
	// blanket dirty check would be confounded — assert on the merge state itself.)
	if _, err := os.Stat(filepath.Join(root, ".git", "MERGE_HEAD")); err == nil {
		t.Error("MERGE_HEAD present; the conflicted merge was not aborted")
	}
	if unmerged := runGit(t, root, "ls-files", "-u"); strings.TrimSpace(unmerged) != "" {
		t.Errorf("unmerged index entries remain after abort:\n%s", unmerged)
	}
	if b := runGit(t, root, "branch", "--list", ib.Branch); strings.TrimSpace(b) == "" {
		t.Errorf("integration branch %q was deleted despite the failed merge", ib.Branch)
	}
	if rec, _ := db.GetIntegrationBranch(parent.ID); rec == nil || rec.Status != "open" {
		t.Errorf("integration record = %+v, want still open after a failed merge", rec)
	}
}

// When the operator has drifted off the integration branch, landing the root
// still commits the root's export onto the integration branch (not the drifted
// branch) so it is included in the merge to main — the export never orphans
// (M1/ADR 0058).
func TestRootLandCommitsExportToIntegrationBranchDespiteDrift(t *testing.T) {
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

	// Operator drifts to an unrelated branch before landing the root.
	runGit(t, root, "checkout", "-b", "scratch")

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

	if cur, _ := srv.repo.CurrentBranch(); cur != def {
		t.Errorf("current branch = %q, want default %q after root landing", cur, def)
	}
	// The root's export must be tracked on the default branch (it rode the
	// integration branch into the merge, not the drifted scratch branch).
	rootExport := ".groundwork/tickets/" + parent.ID + "/ticket.md"
	if tracked := runGit(t, root, "ls-files", rootExport); strings.TrimSpace(tracked) == "" {
		t.Errorf("root export %q not on default branch; it orphaned on the drifted branch", rootExport)
	}
	if b := runGit(t, root, "branch", "--list", ib.Branch); strings.TrimSpace(b) != "" {
		t.Errorf("integration branch %q still exists after landing", ib.Branch)
	}
}

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
