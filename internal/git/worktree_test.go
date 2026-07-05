package git

import (
	"os"
	"path/filepath"
	"testing"
)

// initRepoWithCommit returns a repo with one commit, so worktree/branch ops have
// a base.
func initRepoWithCommit(t *testing.T) (*Repo, string) {
	t.Helper()
	dir := initRepo(t)
	r, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "base.txt"), []byte("base\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := r.Add("base.txt"); err != nil {
		t.Fatal(err)
	}
	if _, err := r.Commit("base commit"); err != nil {
		t.Fatal(err)
	}
	return r, dir
}

func TestWorktreeAddListRemove(t *testing.T) {
	r, dir := initRepoWithCommit(t)
	wt := filepath.Join(dir, "..", "wt-1")

	if err := r.WorktreeAdd(wt, "gw/run/R-1", ""); err != nil {
		t.Fatalf("WorktreeAdd: %v", err)
	}
	if _, err := os.Stat(filepath.Join(wt, "base.txt")); err != nil {
		t.Fatalf("worktree not checked out: %v", err)
	}

	list, err := r.WorktreeList()
	if err != nil {
		t.Fatal(err)
	}
	var found bool
	for _, w := range list {
		if w.Branch == "gw/run/R-1" {
			found = true
		}
	}
	if !found {
		t.Fatalf("gw/run/R-1 not in worktree list: %+v", list)
	}

	if err := r.WorktreeRemove(wt, false); err != nil {
		t.Fatalf("WorktreeRemove: %v", err)
	}
	if _, err := os.Stat(wt); !os.IsNotExist(err) {
		t.Fatalf("worktree dir still present: %v", err)
	}
}

func TestWorktreePruneOrphan(t *testing.T) {
	r, dir := initRepoWithCommit(t)
	wt := filepath.Join(dir, "..", "wt-orphan")
	if err := r.WorktreeAdd(wt, "gw/run/R-2", ""); err != nil {
		t.Fatal(err)
	}
	// Remove the directory out from under git, then prune the stale admin entry.
	if err := os.RemoveAll(wt); err != nil {
		t.Fatal(err)
	}
	if err := r.WorktreePrune(); err != nil {
		t.Fatalf("WorktreePrune: %v", err)
	}
	list, err := r.WorktreeList()
	if err != nil {
		t.Fatal(err)
	}
	for _, w := range list {
		if w.Branch == "gw/run/R-2" {
			t.Fatalf("orphan worktree not pruned: %+v", list)
		}
	}
}

func TestDiffNameOnlyAndMergeSquash(t *testing.T) {
	r, dir := initRepoWithCommit(t)
	base, err := r.HeadCommit()
	if err != nil {
		t.Fatal(err)
	}

	// Make a run branch with a change.
	wt := filepath.Join(dir, "..", "wt-diff")
	if err := r.WorktreeAdd(wt, "gw/run/R-3", base); err != nil {
		t.Fatal(err)
	}
	wtRepo, err := Open(wt)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(wt, "feature.go"), []byte("package x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := wtRepo.AddAll(); err != nil {
		t.Fatal(err)
	}
	if _, err := wtRepo.Commit("add feature"); err != nil {
		t.Fatal(err)
	}

	// The base diff reports only the run branch's own change.
	files, err := r.DiffNameOnly(base, "gw/run/R-3")
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 1 || files[0] != "feature.go" {
		t.Fatalf("DiffNameOnly = %v, want [feature.go]", files)
	}

	// Squash-merge the run branch into the main working tree's index.
	if err := r.MergeSquash("gw/run/R-3"); err != nil {
		t.Fatalf("MergeSquash: %v", err)
	}
	staged, err := r.HasStagedChanges()
	if err != nil || !staged {
		t.Fatalf("after squash: staged=%v err=%v, want staged", staged, err)
	}
}

func TestRunRefRetention(t *testing.T) {
	r, _ := initRepoWithCommit(t)
	head, err := r.HeadCommit()
	if err != nil {
		t.Fatal(err)
	}
	ref := "refs/groundwork/runs/R-9"
	if err := r.UpdateRef(ref, head); err != nil {
		t.Fatalf("UpdateRef: %v", err)
	}
	if err := r.DeleteRef(ref); err != nil {
		t.Fatalf("DeleteRef: %v", err)
	}
}
