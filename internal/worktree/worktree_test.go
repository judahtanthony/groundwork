package worktree

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"groundwork/internal/git"
)

func initRepo(t *testing.T) (*git.Repo, string) {
	t.Helper()
	dir := t.TempDir()
	for _, args := range [][]string{
		{"init"}, {"config", "user.email", "t@example.com"}, {"config", "user.name", "Test"},
	} {
		c := exec.Command("git", args...)
		c.Dir = dir
		if out, err := c.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v: %s", args, err, out)
		}
	}
	r, err := git.Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "base.txt"), []byte("base\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := r.Add("base.txt"); err != nil {
		t.Fatal(err)
	}
	if _, err := r.Commit("base"); err != nil {
		t.Fatal(err)
	}
	return r, dir
}

func TestProvisionAndTeardown(t *testing.T) {
	r, dir := initRepo(t)
	m := NewManager(r, filepath.Join(dir, ".groundwork", "worktrees"))

	p, err := m.Provision("R-1", "")
	if err != nil {
		t.Fatalf("Provision: %v", err)
	}
	if p.Branch != "gw/run/R-1" {
		t.Errorf("branch = %q", p.Branch)
	}
	if _, err := os.Stat(filepath.Join(p.Path, "base.txt")); err != nil {
		t.Fatalf("worktree not provisioned: %v", err)
	}
	if !r.BranchExists("gw/run/R-1") {
		t.Error("run branch missing")
	}

	// Provisioning the same run twice fails rather than clobbering.
	if _, err := m.Provision("R-1", ""); err == nil {
		t.Error("expected error provisioning an existing worktree")
	}

	if err := m.Teardown("R-1", true); err != nil {
		t.Fatalf("Teardown: %v", err)
	}
	if _, err := os.Stat(p.Path); !os.IsNotExist(err) {
		t.Errorf("worktree dir remains: %v", err)
	}
	if r.BranchExists("gw/run/R-1") {
		t.Error("run branch not deleted")
	}
}

func TestRetainKeepsWIPBeforeTeardown(t *testing.T) {
	r, dir := initRepo(t)
	m := NewManager(r, filepath.Join(dir, ".groundwork", "worktrees"))
	p, err := m.Provision("R-2", "")
	if err != nil {
		t.Fatal(err)
	}
	// Commit some WIP on the run branch inside its worktree.
	wt, err := git.Open(p.Path)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(p.Path, "wip.txt"), []byte("wip\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := wt.AddAll(); err != nil {
		t.Fatal(err)
	}
	if _, err := wt.Commit("wip checkpoint"); err != nil {
		t.Fatal(err)
	}

	if err := m.Retain("R-2"); err != nil {
		t.Fatalf("Retain: %v", err)
	}
	if err := m.Teardown("R-2", true); err != nil {
		t.Fatal(err)
	}
	// The WIP chain survives under the run ref even though the branch is gone.
	if err := r.DeleteRef(RunRef("R-2")); err != nil {
		t.Fatalf("run ref was not retained: %v", err)
	}
}

func TestManagerDiffCapturesChanges(t *testing.T) {
	r, dir := initRepo(t)
	m := NewManager(r, filepath.Join(dir, ".groundwork", "worktrees"))
	p, err := m.Provision("R-d", "")
	if err != nil {
		t.Fatal(err)
	}

	// Empty run: no changes captured.
	files, diff, err := m.Diff("R-d", p.Base)
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 0 || diff != "" {
		t.Fatalf("empty run: files=%v diff=%q", files, diff)
	}

	// Add a new file, modify the base file, delete... base.txt removed.
	if err := os.WriteFile(filepath.Join(p.Path, "feature.go"), []byte("package x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(p.Path, "base.txt"), []byte("changed\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	files, diff, err = m.Diff("R-d", p.Base)
	if err != nil {
		t.Fatal(err)
	}
	if !has(files, "feature.go") || !has(files, "base.txt") {
		t.Fatalf("changed files = %v, want feature.go and base.txt", files)
	}
	if diff == "" {
		t.Error("expected a non-empty unified diff")
	}
}

func TestManagerDiffCapturesDeletion(t *testing.T) {
	r, dir := initRepo(t)
	m := NewManager(r, filepath.Join(dir, ".groundwork", "worktrees"))
	p, err := m.Provision("R-del", "")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(filepath.Join(p.Path, "base.txt")); err != nil {
		t.Fatal(err)
	}
	files, _, err := m.Diff("R-del", p.Base)
	if err != nil {
		t.Fatal(err)
	}
	if !has(files, "base.txt") {
		t.Fatalf("deletion not captured: %v", files)
	}
}

func has(s []string, want string) bool {
	for _, v := range s {
		if v == want {
			return true
		}
	}
	return false
}

func TestCheckpointAndSquashLand(t *testing.T) {
	r, dir := initRepo(t)
	m := NewManager(r, filepath.Join(dir, ".groundwork", "worktrees"))
	p, err := m.Provision("R-c", "")
	if err != nil {
		t.Fatal(err)
	}

	// The run produces two files; checkpoint commits them on the run branch.
	if err := os.WriteFile(filepath.Join(p.Path, "feature.go"), []byte("package x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(p.Path, "feature_test.go"), []byte("package x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	sha, err := m.Checkpoint("R-c", "wip")
	if err != nil || sha == "" {
		t.Fatalf("Checkpoint: sha=%q err=%v", sha, err)
	}
	// Nothing new to commit → empty sha, no error.
	sha2, err := m.Checkpoint("R-c", "wip2")
	if err != nil || sha2 != "" {
		t.Fatalf("second Checkpoint: sha=%q err=%v, want empty", sha2, err)
	}

	// Land: squash the run branch into the main working tree's index, then commit
	// one curated landing commit (the integration branch is the checked-out default).
	if err := r.MergeSquash(RunBranch("R-c")); err != nil {
		t.Fatalf("MergeSquash: %v", err)
	}
	landSha, err := r.Commit("Land R-c (squashed)")
	if err != nil {
		t.Fatal(err)
	}
	// The integration branch now has both files in a single squashed commit.
	if _, err := os.Stat(filepath.Join(dir, "feature.go")); err != nil {
		t.Fatalf("squashed file missing on integration branch: %v", err)
	}
	if landSha == "" {
		t.Fatal("expected a landing commit sha")
	}

	// Retain the WIP chain, then tear down the run branch + worktree.
	if err := m.Retain("R-c"); err != nil {
		t.Fatal(err)
	}
	if err := m.Teardown("R-c", true); err != nil {
		t.Fatal(err)
	}
	if r.BranchExists(RunBranch("R-c")) {
		t.Error("run branch not deleted after land")
	}
	// The WIP chain survives under the run ref.
	if err := r.DeleteRef(RunRef("R-c")); err != nil {
		t.Errorf("run ref not retained: %v", err)
	}
}

// TestCheckpointBaseResumesFromBranchThenRef proves a resuming run continues from
// a prior run's checkpoint — its branch while it exists, then the retained run ref
// after teardown — so interrupted in-flight work is not lost (T-0904, ADR 0015).
func TestCheckpointBaseResumesFromBranchThenRef(t *testing.T) {
	r, dir := initRepo(t)
	m := NewManager(r, filepath.Join(dir, ".groundwork", "worktrees"))

	// Prior run produced a checkpoint with in-flight work.
	prior, err := m.Provision("R-1", "")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(prior.Path, "inflight.go"), []byte("package x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := m.Checkpoint("R-1", "wip"); err != nil {
		t.Fatal(err)
	}

	// While the branch exists, resume cuts from it.
	base, ok := m.CheckpointBase("R-1")
	if !ok || base != RunBranch("R-1") {
		t.Fatalf("CheckpointBase = %q,%v; want the run branch", base, ok)
	}

	// The interrupted run is reconciled: its WIP is retained under the ref, the
	// branch and worktree are removed.
	if err := m.Retain("R-1"); err != nil {
		t.Fatal(err)
	}
	if err := m.Teardown("R-1", true); err != nil {
		t.Fatal(err)
	}

	// Now resume cuts from the retained ref instead.
	base, ok = m.CheckpointBase("R-1")
	if !ok || base != RunRef("R-1") {
		t.Fatalf("after teardown CheckpointBase = %q,%v; want the run ref", base, ok)
	}

	// The resuming run's worktree branches from that checkpoint and carries the
	// prior in-flight work forward.
	resumed, err := m.Provision("R-2", base)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(resumed.Path, "inflight.go")); err != nil {
		t.Fatalf("resumed worktree did not carry prior in-flight work: %v", err)
	}

	// No prior checkpoint → not resumable (caller cuts from the integration base).
	if _, ok := m.CheckpointBase("R-nope"); ok {
		t.Error("unexpected checkpoint base for an unknown run")
	}
}

// TestResumedDiffAgainstIntegrationBaseCoversBothRuns proves the git mechanics
// behind review finding #3: a resumed run's diff against the integration base
// includes the interrupted run's files, whereas a diff against the checkpoint
// would omit them.
func TestResumedDiffAgainstIntegrationBaseCoversBothRuns(t *testing.T) {
	r, dir := initRepo(t)
	m := NewManager(r, filepath.Join(dir, ".groundwork", "worktrees"))
	integrationBase, err := r.HeadCommit()
	if err != nil {
		t.Fatal(err)
	}

	// Interrupted run R-1: writes out-of-scope.go, checkpoints, then is reclaimed.
	r1, err := m.Provision("R-1", "")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(r1.Path, "out-of-scope.go"), []byte("package x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := m.Checkpoint("R-1", "wip"); err != nil {
		t.Fatal(err)
	}
	if err := m.Retain("R-1"); err != nil {
		t.Fatal(err)
	}
	if err := m.Teardown("R-1", true); err != nil {
		t.Fatal(err)
	}

	// Resume R-2 from R-1's retained checkpoint and add its own file.
	base, ok := m.CheckpointBase("R-1")
	if !ok {
		t.Fatal("no checkpoint to resume from")
	}
	r2, err := m.Provision("R-2", base)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(r2.Path, "resume.go"), []byte("package x\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Diff against the integration base covers BOTH runs' files.
	files, _, err := m.Diff("R-2", integrationBase)
	if err != nil {
		t.Fatal(err)
	}
	if !has(files, "out-of-scope.go") || !has(files, "resume.go") {
		t.Fatalf("diff vs integration base = %v, want both files", files)
	}

	// Diff against the checkpoint (the buggy base) would omit the interrupted file.
	buggy, _, err := m.Diff("R-2", base)
	if err != nil {
		t.Fatal(err)
	}
	if has(buggy, "out-of-scope.go") {
		t.Fatal("sanity: diff vs checkpoint unexpectedly included the prior file")
	}
}

func TestReconcileReclaimsOrphans(t *testing.T) {
	r, dir := initRepo(t)
	m := NewManager(r, filepath.Join(dir, ".groundwork", "worktrees"))
	if _, err := m.Provision("R-a", ""); err != nil {
		t.Fatal(err)
	}
	if _, err := m.Provision("R-b", ""); err != nil {
		t.Fatal(err)
	}

	// R-a is still active; R-b is orphaned.
	reclaimed, err := m.Reconcile(map[string]bool{"R-a": true})
	if err != nil {
		t.Fatalf("Reconcile: %v", err)
	}
	if len(reclaimed) != 1 || reclaimed[0] != "R-b" {
		t.Fatalf("reclaimed = %v, want [R-b]", reclaimed)
	}
	if _, err := os.Stat(m.Path("R-b")); !os.IsNotExist(err) {
		t.Error("orphan worktree not removed")
	}
	if _, err := os.Stat(m.Path("R-a")); err != nil {
		t.Error("active worktree should be left intact")
	}
}
