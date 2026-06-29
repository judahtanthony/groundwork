package cli

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"slices"
	"testing"
	"time"

	"groundwork/internal/actor"
	"groundwork/internal/approval"
	"groundwork/internal/completion"
	"groundwork/internal/config"
	"groundwork/internal/git"
	"groundwork/internal/policy"
	gwrun "groundwork/internal/run"
	gwruntime "groundwork/internal/runtime"
	"groundwork/internal/scheduler"
	"groundwork/internal/server"
	"groundwork/internal/store/sqlite"
	"groundwork/internal/ticket"
	"groundwork/internal/worktree"
)

// TestCodexAssistedTicketEndToEnd is the Phase 6 capstone (T-1003): it drives a
// real implementation ticket through the Codex runtime end-to-end with a
// deterministic scripted agent standing in for the codex CLI (no codex binary in
// CI). It proves the run is tracked, the agent's work is captured in an isolated
// worktree with a completion summary, and the human landing gate is exercised.
func TestCodexAssistedTicketEndToEnd(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("scripted agent is POSIX-only")
	}
	root := t.TempDir()
	runGit := func(args ...string) {
		c := exec.Command("git", args...)
		c.Dir = root
		if out, err := c.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v: %s", args, err, out)
		}
	}
	runGit("init")
	runGit("config", "user.email", "t@example.com")
	runGit("config", "user.name", "Test")
	runGit("commit", "--allow-empty", "-m", "init")

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

	// The scripted agent: a real implementation step that writes a source file in
	// its worktree (the cwd).
	agent := filepath.Join(root, "agent.sh")
	if err := os.WriteFile(agent, []byte("#!/bin/sh\nprintf 'package feature\\n\\nfunc Hello() string { return \"hi\" }\\n' > feature.go\necho 'implemented feature.go'\n"), 0o755); err != nil {
		t.Fatal(err)
	}

	repo, err := git.Open(root)
	if err != nil {
		t.Fatal(err)
	}
	mgr := worktree.NewManager(repo, proj.WorktreesDir())
	codex := gwruntime.NewCodex(gwruntime.Config{Command: agent, WorktreeRoot: proj.WorktreesDir()}).
		WithExec().WithWorkspace(worktreeProvider{mgr}, resumeBase(db, mgr))
	if codex.Name() != "codex" {
		t.Fatalf("runtime = %q, want codex", codex.Name())
	}

	registry := &actor.Registry{Schema: actor.SchemaVersion, Actors: []actor.Actor{
		{ID: "human.owner", Type: actor.TypeHuman, Roles: []string{"owner"}, Capabilities: actor.Capabilities{WorkTypes: []string{"*"}}},
		{ID: "ai.codex.default", Type: actor.TypeAIAgent, Roles: []string{"coding"}, Capabilities: actor.Capabilities{WorkTypes: []string{"technical_implementation"}}},
	}}
	policies := &policy.Set{Trust: &policy.TrustPolicy{AllowClaim: []policy.Rule{{
		ID:      "codex",
		When:    policy.Match{ActorIDs: []string{"ai.codex.default"}, WorkTypes: []string{"technical_implementation"}, RiskClassAtMost: "high"},
		Actions: []string{"execute"},
	}}}}

	sched := scheduler.New(db, policies, registry, codex, nil, scheduler.Config{
		MaxConcurrency: 1, LeaseTTL: 90 * time.Second, TickInterval: time.Hour,
		Model: "claude-test", RunLogDir: proj.RunsDir(), TicketsDir: proj.TicketsDir(),
	})

	// The implementation ticket.
	tk := &ticket.Ticket{Title: "implement Hello", NodeType: ticket.NodeLeaf, Status: ticket.StatusTodo,
		WorkType: "technical_implementation", Acceptance: []string{"feature.go exists"}}
	if err := db.CreateTicket(tk, "tester"); err != nil {
		t.Fatal(err)
	}

	// Dispatch: the scheduler claims, runs Codex in an isolated worktree, captures
	// the diff, checkpoints, writes a completion summary, and moves to review.
	started, err := sched.Tick(context.Background())
	if err != nil {
		t.Fatalf("Tick: %v", err)
	}
	if started != 1 {
		t.Fatalf("started = %d, want 1", started)
	}
	sched.Wait()

	// The Codex run is tracked, with runtime + model metadata.
	runs, _ := db.ListRunsForTicket(tk.ID)
	if len(runs) != 1 {
		t.Fatalf("runs = %d, want 1", len(runs))
	}
	r := runs[0]
	if r.Runtime != "codex" || r.ActorID != "ai.codex.default" || r.Status != string(gwrun.StatusCompleted) {
		t.Fatalf("run not tracked as expected: %+v", r)
	}

	// The agent's work was captured from the isolated worktree.
	files, _ := db.ChangedFilesForNode(tk.ID)
	if !slices.Contains(files, "feature.go") {
		t.Fatalf("changed files = %v, want feature.go", files)
	}
	if !repo.BranchExists(worktree.RunBranch(r.ID)) {
		t.Error("run branch with the checkpoint is missing")
	}

	// A completion summary was written for the runtime-produced result.
	sum, ok, err := completion.Read(proj.TicketsDir(), tk.ID)
	if err != nil || !ok {
		t.Fatalf("completion summary missing: ok=%v err=%v", ok, err)
	}
	if !slices.Contains(sum.Changed, "feature.go") {
		t.Errorf("summary changed = %v", sum.Changed)
	}

	// The node awaits review.
	if got, _ := db.GetTicket(tk.ID); got.Status != ticket.StatusReview {
		t.Fatalf("status = %s, want review", got.Status)
	}

	// The human landing gate is exercised: requesting a land_to_main on the
	// code change opens a PENDING approval (not auto-approved) for a human.
	srv := server.New(db, proj, "test")
	svc := server.NewApprovalService(db, policies, registry)
	srv.SetApprovals(svc)
	appr, err := svc.RequestLanding(tk.ID, tk.WorkType)
	if err != nil {
		t.Fatal(err)
	}
	if appr.Type != string(approval.TypeLandToMain) || appr.Status != string(approval.StatusPending) {
		t.Fatalf("landing approval not human-gated: %+v", appr)
	}
	// An AI actor cannot decide a human-gated landing.
	if _, err := svc.Decide(appr.ID, approval.StatusApproved, "ai.codex.default", "ship it"); err == nil {
		t.Error("AI actor decided a human-gated landing; gate bypassed")
	}
}
