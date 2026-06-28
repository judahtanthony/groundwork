// Package worktree manages the per-run isolated git worktrees the Phase 6 runtime
// executes in (ADR 0059). Each run gets a worktree at <worktreesDir>/<run-id> on
// its own branch gw/run/<run-id> from a recorded base commit. The WIP chain can
// be retained under refs/groundwork/runs/<run-id> before teardown so no work is
// silently dropped, and orphaned worktrees reconcile against live run records.
package worktree

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"groundwork/internal/git"
)

const (
	branchPrefix = "gw/run/"
	refPrefix    = "refs/groundwork/runs/"
)

// RunBranch is the branch name for a run's worktree.
func RunBranch(runID string) string { return branchPrefix + runID }

// RunRef is the throwaway ref namespace that retains a run's WIP checkpoint chain
// after landing (ADR 0015/0059).
func RunRef(runID string) string { return refPrefix + runID }

// Manager provisions and tears down per-run worktrees under dir.
type Manager struct {
	repo *git.Repo
	dir  string // worktrees root, e.g. .groundwork/worktrees
}

// NewManager builds a Manager rooted at the given worktrees directory.
func NewManager(repo *git.Repo, worktreesDir string) *Manager {
	return &Manager{repo: repo, dir: worktreesDir}
}

// Provisioned describes a live run worktree.
type Provisioned struct {
	RunID  string
	Path   string // absolute worktree path
	Branch string // gw/run/<run-id>
	Base   string // base commit the branch was cut from
}

// Path returns the worktree path for a run.
func (m *Manager) Path(runID string) string { return filepath.Join(m.dir, runID) }

// Provision creates the run's branch at base and checks it out in a fresh
// worktree under the worktrees dir (ADR 0059). base may be a commit or ref; ""
// uses HEAD. It fails if the worktree already exists.
func (m *Manager) Provision(runID, base string) (*Provisioned, error) {
	if runID == "" {
		return nil, fmt.Errorf("worktree: run id is required")
	}
	path := m.Path(runID)
	if _, err := os.Stat(path); err == nil {
		return nil, fmt.Errorf("worktree: %s already exists", path)
	}
	if err := os.MkdirAll(m.dir, 0o755); err != nil {
		return nil, err
	}
	branch := RunBranch(runID)
	if err := m.repo.WorktreeAdd(path, branch, base); err != nil {
		return nil, err
	}
	return &Provisioned{RunID: runID, Path: path, Branch: branch, Base: base}, nil
}

// Retain saves the run branch's current tip under refs/groundwork/runs/<run-id>
// so the WIP checkpoint chain survives teardown (ADR 0015). A no-op when the
// branch does not exist.
func (m *Manager) Retain(runID string) error {
	branch := RunBranch(runID)
	if !m.repo.BranchExists(branch) {
		return nil
	}
	return m.repo.UpdateRef(RunRef(runID), branch)
}

// Teardown removes a run's worktree and deletes its branch. force discards an
// uncommitted/un-landed worktree; callers that must not lose work call Retain
// first (ADR 0059). It prunes stale metadata and tolerates an already-gone tree.
func (m *Manager) Teardown(runID string, force bool) error {
	path := m.Path(runID)
	if _, err := os.Stat(path); err == nil {
		if err := m.repo.WorktreeRemove(path, force); err != nil {
			return err
		}
	}
	_ = m.repo.WorktreePrune()
	branch := RunBranch(runID)
	if m.repo.BranchExists(branch) {
		// The branch is throwaway (its content squash-lands or is retained as a
		// ref); a force delete is intended here.
		if err := m.repo.DeleteBranchForce(branch); err != nil {
			return err
		}
	}
	return nil
}

// List returns the run ids that currently have a worktree directory.
func (m *Manager) List() ([]string, error) {
	entries, err := os.ReadDir(m.dir)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var ids []string
	for _, e := range entries {
		if e.IsDir() {
			ids = append(ids, e.Name())
		}
	}
	return ids, nil
}

// Reconcile removes worktrees whose run is no longer active (ADR 0059 recovery):
// each orphan's WIP is first retained under its run ref so nothing is dropped,
// then its worktree and branch are removed. Returns the reclaimed run ids.
func (m *Manager) Reconcile(active map[string]bool) ([]string, error) {
	ids, err := m.List()
	if err != nil {
		return nil, err
	}
	var reclaimed []string
	for _, id := range ids {
		if active[id] {
			continue
		}
		if err := m.Retain(id); err != nil {
			return reclaimed, err
		}
		if err := m.Teardown(id, true); err != nil {
			return reclaimed, err
		}
		reclaimed = append(reclaimed, id)
	}
	return reclaimed, nil
}

// IsRunWorktreePath reports whether p looks like a managed run worktree path.
func (m *Manager) IsRunWorktreePath(p string) bool {
	return strings.HasPrefix(filepath.Clean(p), filepath.Clean(m.dir)+string(filepath.Separator))
}
