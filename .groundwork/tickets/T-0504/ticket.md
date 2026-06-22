---
id: T-0504
kind: ticket
node_type: leaf
work_type: technical_implementation
title: Implement run checkpoints and landing squash
status: todo
assignee: null
requested_actor: null
priority: null
labels: []
parent: E-0006
depends_on:
    - T-0502
created_at: "2026-06-22T12:20:48Z"
updated_at: "2026-06-22T15:06:06Z"
---

## Problem

_No description recorded._

## Acceptance Criteria

- A run commits WIP checkpoints on its worktree branch, never on main.
- Checkpoints are squashed into the landing commit.
