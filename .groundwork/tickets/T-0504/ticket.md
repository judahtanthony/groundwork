---
id: T-0504
kind: ticket
node_type: leaf
work_type: technical_implementation
title: Implement run checkpoints and landing squash
status: backlog
assignee: null
requested_actor: null
priority: null
labels: []
parent: E-0006
depends_on: []
created_at: ""
updated_at: ""
---

## Problem

_No description recorded._

## Acceptance Criteria

- A run commits WIP checkpoints on its worktree branch, never on main.
- Checkpoints are squashed into the landing commit.
