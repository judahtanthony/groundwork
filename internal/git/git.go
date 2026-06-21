// Package git is a minimal git-landing helper: it stages a pathspec and commits
// on the current branch so a node that Groundwork reports as landed is also
// committed (ADR 0034). It is deliberately small — stage + commit on the current
// branch only. Isolated worktrees, branch creation, WIP checkpoints, squash, and
// resume belong to the Phase 4 runtime (ADR 0027) and are not implemented here.
package git

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

// Repo is a handle to a git work tree rooted at a directory.
type Repo struct {
	root string
}

// ErrNotARepo is returned by Open when root is not inside a git work tree.
var ErrNotARepo = errors.New("not a git work tree")

// Open returns a Repo for root when root is inside a git work tree. Callers
// treat ErrNotARepo as "git landing unavailable" and degrade gracefully (the
// landing is still recorded in the store), so Groundwork works in non-git dirs.
func Open(root string) (*Repo, error) {
	r := &Repo{root: root}
	out, err := r.run("rev-parse", "--is-inside-work-tree")
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrNotARepo, root)
	}
	if strings.TrimSpace(out) != "true" {
		return nil, fmt.Errorf("%w: %s", ErrNotARepo, root)
	}
	return r, nil
}

// Add stages the given paths (which may be absolute or relative to the repo
// root). Pathspecs honoring .gitignore are passed through to `git add`.
func (r *Repo) Add(paths ...string) error {
	if len(paths) == 0 {
		return nil
	}
	args := append([]string{"add", "--"}, paths...)
	_, err := r.run(args...)
	return err
}

// HasStagedChanges reports whether the index differs from HEAD (i.e. there is
// something to commit).
func (r *Repo) HasStagedChanges() (bool, error) {
	// `git diff --cached --quiet` exits 0 when the index matches HEAD and 1 when
	// it differs; any other exit code is a real error.
	cmd := exec.Command("git", "diff", "--cached", "--quiet")
	cmd.Dir = r.root
	err := cmd.Run()
	if err == nil {
		return false, nil
	}
	var ee *exec.ExitError
	if errors.As(err, &ee) && ee.ExitCode() == 1 {
		return true, nil
	}
	return false, fmt.Errorf("git diff --cached: %w", err)
}

// Commit records the staged changes with message and returns the new commit SHA.
// It does not stage anything itself; callers stage first via Add.
func (r *Repo) Commit(message string) (string, error) {
	if _, err := r.run("commit", "-m", message); err != nil {
		return "", err
	}
	sha, err := r.run("rev-parse", "HEAD")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(sha), nil
}

// CurrentBranch returns the checked-out branch name (or "HEAD" when detached).
func (r *Repo) CurrentBranch() (string, error) {
	out, err := r.run("rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(out), nil
}

// WorkingTreeDirty reports whether tracked files have unstaged modifications.
// Landing uses this to warn that part of the work may be unstaged.
func (r *Repo) WorkingTreeDirty() (bool, error) {
	// Exit 1 means the working tree differs from the index.
	cmd := exec.Command("git", "diff", "--quiet")
	cmd.Dir = r.root
	err := cmd.Run()
	if err == nil {
		return false, nil
	}
	var ee *exec.ExitError
	if errors.As(err, &ee) && ee.ExitCode() == 1 {
		return true, nil
	}
	return false, fmt.Errorf("git diff: %w", err)
}

// run executes a git subcommand in the repo root, returning combined output. On
// failure the error includes the command and trimmed output for diagnostics.
func (r *Repo) run(args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = r.root
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	if err := cmd.Run(); err != nil {
		return out.String(), fmt.Errorf("git %s: %v: %s",
			strings.Join(args, " "), err, strings.TrimSpace(out.String()))
	}
	return out.String(), nil
}
