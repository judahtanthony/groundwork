package git

// Worktree primitives for the Phase 6 runtime (ADR 0059): each run executes in an
// isolated `git worktree` on its own `gw/run/<run-id>` branch from a recorded base
// commit, and lands by squashing that branch into the integration branch. These
// are the low-level git calls; the per-run lifecycle and `.groundwork/worktrees`
// path convention live in internal/worktree.

import (
	"bufio"
	"strings"
)

// Worktree is one entry from `git worktree list`.
type Worktree struct {
	Path   string // absolute worktree path
	Head   string // checked-out commit
	Branch string // checked-out branch (refs/heads/... trimmed), or "" when detached
}

// WorktreeAdd creates a new branch at base and checks it out in a fresh worktree
// at path (`git worktree add -b <branch> <path> <base>`). base may be a commit
// SHA or ref; an empty base uses HEAD.
func (r *Repo) WorktreeAdd(path, branch, base string) error {
	args := []string{"worktree", "add", "-b", branch, path}
	if base != "" {
		args = append(args, base)
	}
	_, err := r.run(args...)
	return err
}

// WorktreeRemove removes the worktree at path. With force it discards uncommitted
// changes (`git worktree remove --force`); without, git refuses a dirty worktree,
// so a non-force failure is a real signal that un-landed work would be lost.
func (r *Repo) WorktreeRemove(path string, force bool) error {
	args := []string{"worktree", "remove", path}
	if force {
		args = append(args, "--force")
	}
	_, err := r.run(args...)
	return err
}

// WorktreePrune clears administrative metadata for worktrees whose directory was
// removed out from under git (`git worktree prune`).
func (r *Repo) WorktreePrune() error {
	_, err := r.run("worktree", "prune")
	return err
}

// WorktreeList parses `git worktree list --porcelain`. The main working tree is
// included; callers filter by path/branch as needed.
func (r *Repo) WorktreeList() ([]Worktree, error) {
	out, err := r.run("worktree", "list", "--porcelain")
	if err != nil {
		return nil, err
	}
	var list []Worktree
	var cur Worktree
	flush := func() {
		if cur.Path != "" {
			list = append(list, cur)
		}
		cur = Worktree{}
	}
	sc := bufio.NewScanner(strings.NewReader(out))
	for sc.Scan() {
		line := sc.Text()
		switch {
		case line == "":
			flush()
		case strings.HasPrefix(line, "worktree "):
			cur.Path = strings.TrimPrefix(line, "worktree ")
		case strings.HasPrefix(line, "HEAD "):
			cur.Head = strings.TrimPrefix(line, "HEAD ")
		case strings.HasPrefix(line, "branch "):
			cur.Branch = strings.TrimPrefix(strings.TrimPrefix(line, "branch "), "refs/heads/")
		}
	}
	flush()
	return list, sc.Err()
}

// MergeSquash stages the squashed net change of branch into the index without
// committing (`git merge --squash`), so a single curated landing commit can be
// made by the caller (ADR 0015/0059). A conflict leaves the merge state for the
// caller to abort.
func (r *Repo) MergeSquash(branch string) error {
	_, err := r.run("merge", "--squash", branch)
	return err
}

// DiffNameOnly returns the paths changed on ref relative to base, using the
// merge-base (three-dot) so it reports only ref's own changes even if base has
// advanced (ADR 0059). Empty when there are no changes.
func (r *Repo) DiffNameOnly(base, ref string) ([]string, error) {
	out, err := r.run("diff", "--name-only", base+"..."+ref)
	if err != nil {
		return nil, err
	}
	var files []string
	for _, l := range strings.Split(strings.TrimSpace(out), "\n") {
		if l = strings.TrimSpace(l); l != "" {
			files = append(files, l)
		}
	}
	return files, nil
}

// DiffRange returns the unified diff of ref relative to base (three-dot).
func (r *Repo) DiffRange(base, ref string) (string, error) {
	return r.run("diff", base+"..."+ref)
}

// DiffCachedNameOnly returns the paths in the index that differ from base. Used
// to capture a run's changed-file set from its worktree before checkpoints are
// committed (the index is staged first via AddAll). base "" means HEAD.
func (r *Repo) DiffCachedNameOnly(base string) ([]string, error) {
	if base == "" {
		base = "HEAD"
	}
	out, err := r.run("diff", "--cached", "--name-only", base)
	if err != nil {
		return nil, err
	}
	var files []string
	for _, l := range strings.Split(strings.TrimSpace(out), "\n") {
		if l = strings.TrimSpace(l); l != "" {
			files = append(files, l)
		}
	}
	return files, nil
}

// DiffCached returns the unified diff of the index against base ("" means HEAD).
func (r *Repo) DiffCached(base string) (string, error) {
	if base == "" {
		base = "HEAD"
	}
	return r.run("diff", "--cached", base)
}

// UpdateRef points ref at commit (`git update-ref`). Used to retain a run's WIP
// checkpoint chain under refs/groundwork/runs/<run-id> after landing (ADR 0015).
func (r *Repo) UpdateRef(ref, commit string) error {
	_, err := r.run("update-ref", ref, commit)
	return err
}

// DeleteRef removes ref (`git update-ref -d`). Missing refs are tolerated.
func (r *Repo) DeleteRef(ref string) error {
	_, err := r.run("update-ref", "-d", ref)
	return err
}

// BranchExists reports whether a local branch exists.
func (r *Repo) BranchExists(name string) bool {
	_, err := r.run("show-ref", "--verify", "--quiet", "refs/heads/"+name)
	return err == nil
}

// RefExists reports whether a fully-qualified ref exists (e.g.
// refs/groundwork/runs/<id>), used to resume from a retained checkpoint chain.
func (r *Repo) RefExists(ref string) bool {
	_, err := r.run("show-ref", "--verify", "--quiet", ref)
	return err == nil
}
