package cli

import (
	"bytes"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"groundwork/internal/config"
	"groundwork/internal/git"
	"groundwork/internal/server"
	"groundwork/internal/store/sqlite"
	"groundwork/internal/ticket"
	"groundwork/internal/worktree"
)

// TestTicketLandRoutesRunBackedChildToParent is the regression for the land_to_parent
// gap: a child whose completed run's work lives on gw/run/<id> must, when landed,
// squash that run branch onto the root integration branch — not commit the main
// working tree and orphan the code. It drives the real `gw ticket land <child>` CLI
// path (no --to-parent flag) against a live coordinator, so it exercises the
// auto-route. Against the pre-fix code path (plain land → land_to_main) the run's
// file never reaches the integration branch and the run branch is never cleaned up,
// so this test fails there and passes after the fix.
func TestTicketLandRoutesRunBackedChildToParent(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("git worktree setup is POSIX-only here")
	}
	ctx, srv, db, proj := newLandE2E(t)
	branch, child, runID := seedRunBackedChild(t, srv, db, proj)

	var stdout, stderr bytes.Buffer
	ctx.Stdout = &stdout
	ctx.Stderr = &stderr

	// Plain land — the CLI must auto-route this run-backed child to land_to_parent.
	if err := runTicketLand(ctx, []string{child.ID}); err != nil {
		t.Fatalf("runTicketLand: %v (stderr: %s)", err, stderr.String())
	}
	if !strings.Contains(stdout.String(), branch) {
		t.Errorf("output %q does not mention the integration branch %q", stdout.String(), branch)
	}

	// The run's file is on the integration branch (its work landed, not the main tree).
	if out, err := gitOut(t, proj.Root, "show", branch+":feature.go"); err != nil || !strings.Contains(out, "package feature") {
		t.Fatalf("feature.go not on integration branch %q: out=%q err=%v", branch, out, err)
	}
	// The child is done and the throwaway run branch was cleaned up.
	if got, _ := db.GetTicket(child.ID); got.Status != ticket.StatusDone {
		t.Errorf("child status = %s, want done", got.Status)
	}
	repo, _ := git.Open(proj.Root)
	if repo.BranchExists(worktree.RunBranch(runID)) {
		t.Error("run branch was not cleaned up after landing")
	}
	if _, err := os.Stat(filepath.Join(proj.WorktreesDir(), runID)); !os.IsNotExist(err) {
		t.Errorf("run worktree not torn down: err=%v", err)
	}
}

// TestTicketLandToParentFlagLandsChild covers the explicit `--to-parent` surface:
// it lands the same run-backed child via land_to_parent without relying on
// auto-detection.
func TestTicketLandToParentFlagLandsChild(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("git worktree setup is POSIX-only here")
	}
	ctx, srv, db, proj := newLandE2E(t)
	branch, child, runID := seedRunBackedChild(t, srv, db, proj)

	var stdout, stderr bytes.Buffer
	ctx.Stdout, ctx.Stderr = &stdout, &stderr
	if err := runTicketLand(ctx, []string{"--to-parent", child.ID}); err != nil {
		t.Fatalf("runTicketLand --to-parent: %v (stderr: %s)", err, stderr.String())
	}
	if out, err := gitOut(t, proj.Root, "show", branch+":feature.go"); err != nil || !strings.Contains(out, "package feature") {
		t.Fatalf("feature.go not on integration branch: out=%q err=%v", out, err)
	}
	repo, _ := git.Open(proj.Root)
	if repo.BranchExists(worktree.RunBranch(runID)) {
		t.Error("run branch was not cleaned up after landing")
	}
}

// newLandE2E stands up a git-backed project, a live coordinator over it, and a CLI
// Context whose config points at that coordinator.
func newLandE2E(t *testing.T) (*Context, *server.Server, *sqlite.DB, *config.Project) {
	t.Helper()
	root := t.TempDir()
	gitInit(t, root)

	gw := filepath.Join(root, config.GroundworkDir)
	if err := os.MkdirAll(gw, 0o755); err != nil {
		t.Fatal(err)
	}
	db, err := sqlite.Open(filepath.Join(gw, "state.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	if err := db.Migrate(); err != nil {
		t.Fatal(err)
	}
	cfg := config.Defaults()
	proj := &config.Project{Root: root, Config: &cfg}
	db.SetExportDir(proj.TicketsDir())

	srv := server.New(db, proj, "test")
	ts := httptest.NewServer(srv.Handler())
	t.Cleanup(ts.Close)

	// Point the on-disk config at the live coordinator so requireCoordinator resolves it.
	cfg.Server.Addr = ts.Listener.Addr().String()
	data, err := config.Marshal(&cfg)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(proj.ConfigPath(), data, 0o644); err != nil {
		t.Fatal(err)
	}
	return &Context{RootFlag: root}, srv, db, proj
}

// seedRunBackedChild creates a root with an open integration branch and a child
// leaf whose completed run's work sits on a gw/run/<id> branch (feature.go),
// mirroring the state left by a real Codex run before landing.
func seedRunBackedChild(t *testing.T, srv *server.Server, db *sqlite.DB, proj *config.Project) (branch string, child *ticket.Ticket, runID string) {
	t.Helper()
	parent := &ticket.Ticket{Title: "feature root", NodeType: ticket.NodeComposite, Status: ticket.StatusTodo, WorkType: "technical_design"}
	if err := db.CreateTicket(parent, "tester"); err != nil {
		t.Fatal(err)
	}
	branch = "gw/root/" + parent.ID + "-feature-root"
	base, err := gitOut(t, proj.Root, "rev-parse", "HEAD")
	if err != nil {
		t.Fatal(err)
	}
	base = strings.TrimSpace(base)
	gitRun(t, proj.Root, "branch", branch, base)
	if err := db.RecordIntegrationBranch(parent.ID, branch, base); err != nil {
		t.Fatal(err)
	}

	child = &ticket.Ticket{ParentID: parent.ID, Title: "implement hello", NodeType: ticket.NodeLeaf, Status: ticket.StatusInProgress, WorkType: "technical_implementation"}
	if err := db.CreateTicket(child, "tester"); err != nil {
		t.Fatal(err)
	}

	runID = "R-500"
	if _, err := db.Exec(`INSERT INTO runs
		(id, ticket_id, actor_id, actor_snapshot_json, mode, runtime, model, status, workspace_path, started_at, updated_at)
		VALUES (?,?,?,?,?,?,?,?,?,?,?)`,
		runID, child.ID, "ai.codex.default", "{}", "implementation", "codex", "m", "completed", "", "2026-06-28T10:00:00Z", "2026-06-28T10:00:00Z"); err != nil {
		t.Fatal(err)
	}
	repo, err := git.Open(proj.Root)
	if err != nil {
		t.Fatal(err)
	}
	mgr := worktree.NewManager(repo, proj.WorktreesDir())
	p, err := mgr.Provision(runID, branch)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(p.Path, "feature.go"), []byte("package feature\n\nfunc Hello() string { return \"hi\" }\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := mgr.Checkpoint(runID, "wip"); err != nil {
		t.Fatal(err)
	}
	return branch, child, runID
}

func gitInit(t *testing.T, dir string) {
	t.Helper()
	gitRun(t, dir, "init")
	gitRun(t, dir, "config", "user.email", "t@example.com")
	gitRun(t, dir, "config", "user.name", "Test")
	gitRun(t, dir, "commit", "--allow-empty", "-m", "init")
}

func gitRun(t *testing.T, dir string, args ...string) {
	t.Helper()
	if _, err := gitOut(t, dir, args...); err != nil {
		t.Fatalf("git %v: %v", args, err)
	}
}

func gitOut(t *testing.T, dir string, args ...string) (string, error) {
	t.Helper()
	c := exec.Command("git", args...)
	c.Dir = dir
	out, err := c.CombinedOutput()
	return string(out), err
}
