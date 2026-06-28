---
id: T-1059
kind: ticket
node_type: null
work_type: technical_implementation
title: Make durable ticket state file-authoritative
status: todo
assignee: null
requested_actor: null
priority: 0.5
labels:
    - async-agents
    - source-of-truth
parent: T-1052
depends_on:
    - T-1053
created_at: "2026-06-24T18:05:11Z"
updated_at: "2026-06-24T18:05:11Z"
---

## Problem

Implement ADR 0053 by making durable ticket/work-tree mutations write through to the filesystem source of truth, or to a durable replay record, before reporting durable success. SQLite should be rebuildable projection/cache for ticket state, dependencies, durable decision records, and live queue projections.

## Acceptance Criteria

- Create, edit, transition, dependency link/unlink, reparent, decompose, escalate, decision/input/approval, rework, and recovery-state mutations update durable files or a durable replay record before success.
- Deleting .groundwork/state.sqlite and rebuilding from files preserves ticket records, dependencies, statuses, blockers, pending durable requests, and decision records after each durable mutation.
- Startup reconciliation detects SQLite/file divergence and surfaces recovery_needed for unexported durable mutations instead of silently trusting SQLite.
- Tests cover DB purge/rebuild after representative durable mutations.
