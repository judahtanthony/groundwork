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
