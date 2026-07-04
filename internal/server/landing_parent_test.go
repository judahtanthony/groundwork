package server

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"groundwork/internal/approval"
	"groundwork/internal/envelope"
	"groundwork/internal/git"
	"groundwork/internal/store/sqlite"
	"groundwork/internal/ticket"
	"groundwork/internal/worktree"
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

// A child with a failing validation result must not land to its integration
// branch: land_to_parent enforces the same validation gate as land_to_main, so a
// lighter landing level cannot be used to commit work over a red check (H2/ADR 0058).
func TestLandToParentBlockedByFailingValidation(t *testing.T) {
	srv, db, root := newGitServer(t)
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

	child := &ticket.Ticket{ParentID: parent.ID, Title: "child", NodeType: ticket.NodeLeaf, Status: ticket.StatusInProgress, WorkType: "technical_implementation"}
	if err := db.CreateTicket(child, "tester"); err != nil {
		t.Fatal(err)
	}
	if _, err := db.RecordValidation(sqlite.ValidationResult{TicketID: child.ID, Name: "go test", Status: sqlite.ValidationFail}); err != nil {
		t.Fatal(err)
	}

	before := runGit(t, root, "rev-parse", "HEAD")
	if _, err := srv.LandToParent(child.ID); !errors.Is(err, sqlite.ErrValidationGate) {
		t.Fatalf("LandToParent err = %v, want ErrValidationGate", err)
	}
	if after := runGit(t, root, "rev-parse", "HEAD"); after != before {
		t.Error("HEAD advanced despite a failing validation; gate did not block the commit")
	}
	if c, _ := db.GetTicket(child.ID); c.Status == ticket.StatusDone {
		t.Error("child marked done despite a failing validation")
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

// landRouteFor tells the CLI to route a run-backed child to land_to_parent (so its
// run branch is squashed onto the integration branch) while a root and an
// unparented leaf keep the land_to_main path (ADR 0058). Without this, plain
// `gw ticket land <child>` would commit the main working tree and orphan the run.
func TestLandRouteForRunBackedChildIsParent(t *testing.T) {
	srv, db, root := newGitServer(t)
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

	child := &ticket.Ticket{ParentID: parent.ID, Title: "child", NodeType: ticket.NodeLeaf, Status: ticket.StatusInProgress, WorkType: "technical_implementation"}
	if err := db.CreateTicket(child, "tester"); err != nil {
		t.Fatal(err)
	}
	// A completed run with a live run branch carrying the child's work.
	runID := "R-200"
	if _, err := db.Exec(`INSERT INTO runs
		(id, ticket_id, actor_id, actor_snapshot_json, mode, runtime, model, status, workspace_path, started_at, updated_at)
		VALUES (?,?,?,?,?,?,?,?,?,?,?)`,
		runID, child.ID, "ai.codex.default", "{}", "implementation", "codex", "m", "completed", "", "2026-06-28T10:00:00Z", "2026-06-28T10:00:00Z"); err != nil {
		t.Fatal(err)
	}
	mgr := worktree.NewManager(srv.repo, filepath.Join(root, ".groundwork", "worktrees"))
	p, err := mgr.Provision(runID, ib.Branch)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(p.Path, "feature.go"), []byte("package x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := mgr.Checkpoint(runID, "wip"); err != nil {
		t.Fatal(err)
	}

	route, branch, runBranch, err := srv.landRouteFor(child.ID)
	if err != nil {
		t.Fatalf("landRouteFor(child): %v", err)
	}
	if route != "parent" || branch != ib.Branch || !runBranch {
		t.Errorf("child route = %q branch=%q runBranch=%v, want parent %q true", route, branch, runBranch, ib.Branch)
	}

	// The root owns its integration branch, so it lands to main, not to itself.
	if route, _, _, err := srv.landRouteFor(parent.ID); err != nil || route != "main" {
		t.Errorf("root route = %q (err %v), want main", route, err)
	}

	// An unparented leaf with no integration chain lands to main.
	orphan := &ticket.Ticket{Title: "orphan", NodeType: ticket.NodeLeaf, Status: ticket.StatusInProgress, WorkType: "technical_implementation"}
	if err := db.CreateTicket(orphan, "tester"); err != nil {
		t.Fatal(err)
	}
	if route, _, _, err := srv.landRouteFor(orphan.ID); err != nil || route != "main" {
		t.Errorf("orphan route = %q (err %v), want main", route, err)
	}
}

// Concurrent land_to_parent calls must not corrupt the single shared working tree
// (review finding #5): they are serialized by the repo mutex. Run under -race.
func TestConcurrentLandToParentSerialized(t *testing.T) {
	srv, db, root := newGitServer(t)
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
	repo, err := git.Open(root)
	if err != nil {
		t.Fatal(err)
	}
	mgr := worktree.NewManager(repo, filepath.Join(root, ".groundwork", "worktrees"))

	// Two children, each with its own run worktree + checkpoint touching a distinct file.
	var children []string
	for i, name := range []string{"alpha", "beta"} {
		c := &ticket.Ticket{ParentID: parent.ID, Title: name, NodeType: ticket.NodeLeaf, Status: ticket.StatusInProgress, WorkType: "technical_implementation"}
		if err := db.CreateTicket(c, "tester"); err != nil {
			t.Fatal(err)
		}
		runID := "R-" + name
		if _, err := db.Exec(`INSERT INTO runs (id, ticket_id, actor_id, actor_snapshot_json, mode, runtime, model, status, workspace_path, started_at, updated_at)
			VALUES (?,?,?,?,?,?,?,?,?,?,?)`, runID, c.ID, "ai.codex.default", "{}", "implementation", "codex", "m", "completed", "",
			"2026-06-28T1"+string(rune('0'+i))+":00:00Z", "2026-06-28T10:00:00Z"); err != nil {
			t.Fatal(err)
		}
		p, err := mgr.Provision(runID, ib.Branch)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(p.Path, name+".go"), []byte("package x\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		if _, err := mgr.Checkpoint(runID, "wip"); err != nil {
			t.Fatal(err)
		}
		children = append(children, c.ID)
	}

	// Land both concurrently.
	errs := make(chan error, len(children))
	for _, id := range children {
		go func(id string) {
			_, err := srv.LandToParent(id)
			errs <- err
		}(id)
	}
	for range children {
		if err := <-errs; err != nil {
			t.Fatalf("concurrent LandToParent: %v", err)
		}
	}
	// Both children landed; both files are on the integration branch.
	for _, id := range children {
		if c, _ := db.GetTicket(id); c.Status != ticket.StatusDone {
			t.Errorf("%s status = %s, want done", id, c.Status)
		}
	}
	for _, f := range []string{"alpha.go", "beta.go"} {
		if _, err := os.Stat(filepath.Join(root, f)); err != nil {
			t.Errorf("%s not landed on the integration branch: %v", f, err)
		}
	}
}

// A squash conflict during land_to_parent must leave the child NOT done, so it can
// be re-landed after the conflict is resolved (review finding #6).
func TestLandToParentSquashConflictLeavesNodeNotDone(t *testing.T) {
	srv, db, root := newGitServer(t)
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

	// A shared file with a common base on the integration branch.
	if err := os.WriteFile(filepath.Join(root, "shared.go"), []byte("base\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	runGit(t, root, "add", "shared.go")
	runGit(t, root, "commit", "-m", "base shared.go")

	child := &ticket.Ticket{ParentID: parent.ID, Title: "child", NodeType: ticket.NodeLeaf, Status: ticket.StatusInProgress, WorkType: "technical_implementation"}
	if err := db.CreateTicket(child, "tester"); err != nil {
		t.Fatal(err)
	}
	runID := "R-conflict"
	if _, err := db.Exec(`INSERT INTO runs (id, ticket_id, actor_id, actor_snapshot_json, mode, runtime, model, status, workspace_path, started_at, updated_at)
		VALUES (?,?,?,?,?,?,?,?,?,?,?)`, runID, child.ID, "ai.codex.default", "{}", "implementation", "codex", "m", "completed", "",
		"2026-06-28T10:00:00Z", "2026-06-28T10:00:00Z"); err != nil {
		t.Fatal(err)
	}
	repo, _ := git.Open(root)
	mgr := worktree.NewManager(repo, filepath.Join(root, ".groundwork", "worktrees"))
	p, err := mgr.Provision(runID, ib.Branch)
	if err != nil {
		t.Fatal(err)
	}
	// The run changes shared.go one way...
	if err := os.WriteFile(filepath.Join(p.Path, "shared.go"), []byte("run version\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := mgr.Checkpoint(runID, "wip"); err != nil {
		t.Fatal(err)
	}
	// ...and the integration branch advances with a conflicting change.
	if err := os.WriteFile(filepath.Join(root, "shared.go"), []byte("integration version\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	runGit(t, root, "add", "shared.go")
	runGit(t, root, "commit", "-m", "integration shared.go")

	if _, err := srv.LandToParent(child.ID); err == nil {
		t.Fatal("expected a squash conflict error")
	}
	// The child is NOT done — it can be re-landed once the conflict is resolved.
	if c, _ := db.GetTicket(child.ID); c.Status == ticket.StatusDone {
		t.Fatalf("child marked done despite a squash conflict")
	}
	// The reset restored the integration branch's version — no conflict markers left
	// in the tracked file (the tree is not stuck mid-conflict).
	got, err := os.ReadFile(filepath.Join(root, "shared.go"))
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "integration version\n" {
		t.Fatalf("shared.go not reset after failed squash: %q", got)
	}
}

// A child that ran in an isolated worktree lands by squashing its gw/run/<id>
// branch into the integration branch — one curated commit, the run branch torn
// down, and the WIP chain retained under the run ref (T-0504, ADR 0015/0059).
func TestLandToParentSquashesRunBranch(t *testing.T) {
	srv, db, root := newGitServer(t)
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

	child := &ticket.Ticket{ParentID: parent.ID, Title: "child", NodeType: ticket.NodeLeaf, Status: ticket.StatusInProgress, WorkType: "technical_implementation"}
	if err := db.CreateTicket(child, "tester"); err != nil {
		t.Fatal(err)
	}

	// Simulate a Codex run: a run row + an isolated worktree on gw/run/<id> from the
	// integration branch, with a checkpoint commit carrying the run's work.
	runID := "R-100"
	if _, err := db.Exec(`INSERT INTO runs
		(id, ticket_id, actor_id, actor_snapshot_json, mode, runtime, model, status, workspace_path, started_at, updated_at)
		VALUES (?,?,?,?,?,?,?,?,?,?,?)`,
		runID, child.ID, "ai.codex.default", "{}", "implementation", "codex", "m", "completed", "", "2026-06-28T10:00:00Z", "2026-06-28T10:00:00Z"); err != nil {
		t.Fatal(err)
	}
	repo, err := git.Open(root)
	if err != nil {
		t.Fatal(err)
	}
	mgr := worktree.NewManager(repo, filepath.Join(root, ".groundwork", "worktrees"))
	p, err := mgr.Provision(runID, ib.Branch)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(p.Path, "feature.go"), []byte("package x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := mgr.Checkpoint(runID, "wip"); err != nil {
		t.Fatal(err)
	}

	before := runGit(t, root, "rev-parse", ib.Branch)
	if _, err := srv.LandToParent(child.ID); err != nil {
		t.Fatal(err)
	}

	// One squashed commit advanced the integration branch and the run's file landed.
	if after := runGit(t, root, "rev-parse", ib.Branch); after == before {
		t.Error("integration branch did not advance")
	}
	if _, err := os.Stat(filepath.Join(root, "feature.go")); err != nil {
		t.Errorf("squashed run file not on the integration branch: %v", err)
	}
	if c, _ := db.GetTicket(child.ID); c.Status != ticket.StatusDone {
		t.Errorf("child status = %s, want done", c.Status)
	}
	// The throwaway run branch is gone; its WIP chain is retained under the run ref.
	if repo.BranchExists(worktree.RunBranch(runID)) {
		t.Error("run branch not torn down after land")
	}
	if err := repo.DeleteRef(worktree.RunRef(runID)); err != nil {
		t.Errorf("run WIP chain not retained under ref: %v", err)
	}
}
