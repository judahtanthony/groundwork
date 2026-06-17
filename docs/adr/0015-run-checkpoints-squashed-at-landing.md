# ADR 0015: Run Checkpoints Squashed At Landing

Status: Accepted

## Context

Worktree contents are tier-1 runtime state, ignored by default (ADR 0007, ADR 0012), and `recovery.md` admits active runs "may lose model-internal context." But the uncommitted *code* in the worktree is also unprotected, so a crash mid-run can silently lose substantial agent work. The `checkpoints/` directory in `run-logs.md` was listed but never defined. The fix must not clutter `main` or complicate parallel-worktree integration.

## Decision

A run periodically **checkpoints** by committing its work-in-progress on the run's own worktree branch (for example `agent/<node-id>`), never on `main`. Recovery resumes an interrupted run from the last checkpoint commit rather than only from node metadata, turning "may lose work" into "resume from checkpoint."

Constraints are satisfied because checkpoints never leave the worktree:

- Checkpoints are WIP commits on the isolated worktree branch; they are invisible to other worktrees, so integration is unchanged (it still diffs the worktree against base).
- At **landing**, the WIP checkpoints are **squashed** into the curated landing commit(s); `main` sees one clean commit and the WIP chain is dropped (retained only in ignored run logs).
- Optionally, checkpoints live under a throwaway ref namespace (`refs/groundwork/runs/<run-id>`) so they do not clutter the branch list.

This leans on git, an existing local dependency (ADR 0008), and makes the worktree itself the durable in-flight artifact while keeping `main` history clean.

## Consequences

`run-logs.md` defines the `checkpoints/` mechanism, `runtime-model.md` adds checkpointing to the run lifecycle, and `recovery.md` resumes from the last checkpoint. Landing must squash WIP commits so the durable-vs-runtime boundary (ADR 0012) holds: checkpoints are recovery aids that disappear at the ratification gate.
