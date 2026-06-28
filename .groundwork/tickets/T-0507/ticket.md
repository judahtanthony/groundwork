---
id: T-0507
kind: ticket
node_type: leaf
work_type: technical_implementation
title: Capture run changed-file set from worktree and feed gate inputs
status: todo
assignee: null
requested_actor: null
priority: 0.66
labels:
    - phase-6
    - codex-runtime
parent: E-0006
depends_on:
    - T-0502
created_at: "2026-06-28T22:00:00Z"
updated_at: "2026-06-28T22:00:00Z"
---

## Problem

ADR 0059: the changed-file set that feeds gates — validation-template selection, envelope
file-scope (`envelopeScopeAllows`), and the escalation triggers (E-0012) — must come from
the run worktree's diff against its base commit, not the shared git index
(`s.repo.StagedDiff()`). Capture the run diff on the run record and expose it as the
authoritative changed-file set for runtime-produced work.

## Acceptance Criteria

- A completed run records its changed-file set (`git diff --name-only <base>...gw/run/<id>`)
  and full diff as run evidence.
- The coordinator exposes the run's changed-file set to gate inputs (the diff source for
  envelope scope and escalation checks).
- The shared-index land-preview path is unchanged for human/manual landing.
- Tests cover a run that adds, modifies, and deletes files, and an empty (no-change) run.
