---
id: T-0504
kind: ticket
node_type: leaf
work_type: technical_implementation
title: Implement run checkpoints and landing squash
status: todo
assignee: null
requested_actor: null
priority: 0.58
labels: []
parent: E-0006
depends_on:
    - T-0502
created_at: "2026-06-22T12:20:48Z"
updated_at: "2026-06-22T15:06:06Z"
---

## Problem

ADR 0015/0059: a run periodically commits WIP checkpoints on its `gw/run/<run-id>` branch (never on main or the integration branch). At `land_to_parent` the run branch's net diff is squashed into one curated commit on `gw/root/<id>-<slug>`; the WIP chain is retained only under `refs/groundwork/runs/<run-id>`.

## Acceptance Criteria

- A run commits WIP checkpoints on its worktree branch, never on main.
- Checkpoints are squashed into the landing commit.
