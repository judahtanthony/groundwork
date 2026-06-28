---
id: T-0502
kind: ticket
node_type: leaf
work_type: technical_implementation
title: Launch Codex in isolated worktree
status: todo
assignee: null
requested_actor: null
priority: 0.68
labels: []
parent: E-0006
depends_on:
    - T-0506
created_at: "2026-06-22T12:20:48Z"
updated_at: "2026-06-22T15:06:06Z"
---

## Problem

ADR 0059: run the Codex adapter with its working directory set to the run's isolated worktree (`gw/run/<run-id>`, provisioned by T-0506). Validate workspace containment so a misconfigured run cannot escape its worktree, and tear the worktree down at run end.

## Acceptance Criteria

- Runtime validates workspace containment.
- Codex command runs with configured cwd.
