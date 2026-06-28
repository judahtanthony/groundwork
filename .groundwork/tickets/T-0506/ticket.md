---
id: T-0506
kind: ticket
node_type: leaf
work_type: technical_implementation
title: Add per-run git worktree primitives and lifecycle
status: todo
assignee: null
requested_actor: null
priority: 0.72
labels:
    - phase-6
    - codex-runtime
parent: E-0006
depends_on:
    - T-0501
created_at: "2026-06-28T22:00:00Z"
updated_at: "2026-06-28T22:00:00Z"
---

## Problem

ADR 0059 makes each run execute in its own git worktree on branch `gw/run/<run-id>`
branched from the run's base commit, under `.groundwork/worktrees/<run-id>/` (git-ignored).
`internal/git` currently has no worktree primitives. Add them and the provision/teardown
lifecycle so the Codex adapter (T-0502), checkpoints/squash (T-0504), and diff capture
(T-0507) have an isolated tree to operate in.

## Acceptance Criteria

- `internal/git` gains WorktreeAdd, WorktreeRemove, WorktreeList/prune, MergeSquash, and
  base-diff helpers (paths and names), all rooted at the repo.
- A run provisions `gw/run/<run-id>` from its integration base under
  `.groundwork/worktrees/<run-id>/` and records the path on the run.
- Teardown removes the worktree on completion/cancel; orphaned worktrees reconcile against
  run records (`git worktree list`) and never silently drop un-landed work.
- Failure-path tests cover add/remove, prune of an orphan, and base-diff on a dirty tree.
