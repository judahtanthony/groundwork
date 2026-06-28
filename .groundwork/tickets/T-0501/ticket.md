---
id: T-0501
kind: ticket
node_type: leaf
work_type: technical_implementation
title: Create runtime interface and Codex adapter shell
status: todo
assignee: null
requested_actor: null
priority: 0.78
labels: []
parent: E-0006
depends_on: []
created_at: "2026-06-22T12:20:48Z"
updated_at: "2026-06-22T15:06:06Z"
---

## Problem

Replace the records-only stub runtime with a real Codex adapter shell behind the existing `runtime.Runtime` seam (ADR 0027). The adapter is selectable by config and launches with the actor configuration the coordinator selected. This ticket is the shell — process launch and worktree wiring follow in T-0502/T-0506.

## Acceptance Criteria

- Runtime interface compiles.
- Codex adapter is selectable by config.
- Runtime launch accepts actor configuration selected by the coordinator.
