# ADR 0034: Minimal Git-Landing — Groundwork Owns the Landing Commit

Status: Accepted

## Context

All durable state lives in git. `state-model.md`'s master test — "could Groundwork
rebuild this from the files plus git, or is its loss irreversible?" — makes docs,
code, policies, SOPs, and the ticket exports under `.groundwork/tickets/**/ticket.md`
committed truth, while `state.sqlite` and `runs/`/`approvals/` are ignored and
recomputable.

The M2 landing gate ([ADR 0028](0028-gate-evaluation-engine.md)) opens a
`land_to_main` approval but does not itself commit to git;
[ADR 0027](0027-run-lifecycle-and-checkpoint-records.md) places real git operations
(per-run worktree branches, WIP checkpoints, squash, resume) in Phase 4. So a node
that gw reports as `landed`/`done` is not yet *committed* — the system of record can
drift from the durable truth, and a human is relied on to remember `git commit`.

## Decision

**gw owns the landing commit, via a minimal `internal/git`.** After the
`land_to_main` approval passes, `gw ticket land <id>` stages a ticket-scoped pathspec
plus the regenerated ticket export (now `status: done`) and commits on the **current
branch** with a ticket-referencing message, returning the commit SHA (recorded on the
approval/audit trail). The ticket moves to `done` and the commit happens atomically:
gw `done` ⇔ git commit, with the human's change and the updated export in one commit.

The capability is deliberately minimal — **stage + commit on the current branch
only**. It excludes isolated worktrees, branch creation, WIP checkpoints, squash, and
resume, all of which remain Phase 4
([ADR 0027](0027-run-lifecycle-and-checkpoint-records.md)) because they serve
autonomous agent execution, not landing. Guard: if the working tree carries changes
outside the ticket pathspec, gw warns/refuses rather than committing unrelated work.

## Consequences

Landing becomes atomic and gw's recorded state stays consistent with git, so the
self-hosting story needs no manual commit step. This introduces `internal/git`
earlier than the `architecture-map.md` phasing (which pencils it into Phase 4) — a
documented deviation scoped strictly to landing. Phase 4's worktree runtime builds
the richer git layer (checkpoint commits on a run branch, squash-to-landing) *on top*
of this; the landing-commit semantics defined here are the target those checkpoints
squash into. For human-performed M3 work the human edits in the main working tree (no
isolation is needed for a single local human), the pathspec is the ticket's
declared/changed files plus the export, and the dirty-tree guard prevents accidental
capture of unrelated edits.
