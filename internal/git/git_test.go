package git

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func initRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	for _, args := range [][]string{
		{"init"},
		{"config", "user.email", "t@example.com"},
		{"config", "user.name", "Test"},
	} {
		c := exec.Command("git", args...)
		c.Dir = dir
		if out, err := c.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v: %s", args, err, out)
		}
	}
	return dir
}

func TestOpenRejectsNonRepo(t *testing.T) {
	if _, err := Open(t.TempDir()); !errors.Is(err, ErrNotARepo) {
		t.Fatalf("err = %v, want ErrNotARepo", err)
	}
}

func TestAddCommitCycle(t *testing.T) {
	dir := initRepo(t)
	r, err := Open(dir)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}

	if staged, err := r.HasStagedChanges(); err != nil || staged {
		t.Fatalf("clean repo: staged=%v err=%v, want false/nil", staged, err)
	}

	if err := os.WriteFile(filepath.Join(dir, "a.txt"), []byte("hi\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := r.Add("a.txt"); err != nil {
		t.Fatalf("Add: %v", err)
	}
	if staged, err := r.HasStagedChanges(); err != nil || !staged {
		t.Fatalf("after Add: staged=%v err=%v, want true/nil", staged, err)
	}

	sha, err := r.Commit("first commit")
	if err != nil {
		t.Fatalf("Commit: %v", err)
	}
	if len(sha) < 7 {
		t.Fatalf("sha = %q, want a full hash", sha)
	}
	if staged, _ := r.HasStagedChanges(); staged {
		t.Errorf("index not clean after commit")
	}

	br, err := r.CurrentBranch()
	if err != nil || br == "" || br == "HEAD" {
		t.Fatalf("CurrentBranch = %q (err %v), want a named branch", br, err)
	}
}
