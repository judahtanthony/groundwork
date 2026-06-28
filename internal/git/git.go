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

// StagedDiff returns the staged diff (index vs HEAD) — the change set a landing
// commit would record. Empty when nothing is staged.
func (r *Repo) StagedDiff() (string, error) {
	return r.run("diff", "--cached")
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

// HeadCommit returns the current HEAD commit SHA.
func (r *Repo) HeadCommit() (string, error) {
	out, err := r.run("rev-parse", "HEAD")
	return strings.TrimSpace(out), err
}

// CreateAndCheckout creates a new branch at HEAD and checks it out (git checkout
// -b). Used to start a root integration branch from the default branch (ADR 0058).
func (r *Repo) CreateAndCheckout(name string) error {
	_, err := r.run("checkout", "-b", name)
	return err
}

// Checkout switches to an existing branch.
func (r *Repo) Checkout(name string) error {
	_, err := r.run("checkout", name)
	return err
}

// DefaultBranch returns the repo's main line: "main" if it exists, else
// "master", else "" — the merge target for root land_to_main (ADR 0058).
func (r *Repo) DefaultBranch() string {
	for _, b := range []string{"main", "master"} {
		if _, err := r.run("show-ref", "--verify", "--quiet", "refs/heads/"+b); err == nil {
			return b
		}
	}
	return ""
}

// MergeNoFF merges branch into the current branch with a merge commit
// (--no-ff), preserving the feature boundary (ADR 0058).
func (r *Repo) MergeNoFF(branch, message string) error {
	_, err := r.run("merge", "--no-ff", "-m", message, branch)
	return err
}

// MergeAbort aborts an in-progress merge, restoring the work tree and index to
// the pre-merge state (git merge --abort). Used to recover from a conflicted
// root land_to_main so a failed merge never leaves the tree mid-conflict (ADR 0058).
func (r *Repo) MergeAbort() error {
	_, err := r.run("merge", "--abort")
	return err
}

// DeleteBranch removes a branch with the safe `git branch -d`, which refuses to
// delete a branch that is not fully merged. Callers delete only after a
// successful merge (ADR 0058), so a refusal here is a real signal that work would
// be lost — it surfaces as a loud error rather than a silent force-delete.
func (r *Repo) DeleteBranch(name string) error {
	_, err := r.run("branch", "-d", name)
	return err
}

// AddAll stages every change in the work tree (git add -A): modified, deleted,
// and new files, honoring .gitignore. It is the `git commit -a`-style convenience
// for landing (ADR 0034).
func (r *Repo) AddAll() error {
	_, err := r.run("add", "-A")
	return err
}

// HasUncommitted reports whether the work tree has any uncommitted change —
// staged, unstaged, or untracked — honoring .gitignore. Landing uses it to decide
// whether to offer to stage everything when the index is empty.
func (r *Repo) HasUncommitted() (bool, error) {
	out, err := r.run("status", "--porcelain")
	if err != nil {
		return false, err
	}
	return strings.TrimSpace(out) != "", nil
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
