---
id: T-1071
kind: epic
node_type: null
work_type: technical_design
title: 'Phase 6: Durable handoff and Codex runtime'
status: done
assignee: null
requested_actor: null
priority: 0.54
labels:
    - phase-6
    - revised-plan
parent: G-0001
depends_on:
    - T-1066
created_at: "2026-06-24T22:30:12Z"
updated_at: "2026-06-24T22:30:12Z"
---

## Problem

Phase 6 implements the runtime/backend execution substrate after operator UI and bounded autonomy have reduced manual approval overhead. It includes durable async handoff, filesystem-authoritative ticket state, and the real Codex runtime in isolated worktrees with events, transcripts, checkpoints, and squash landing.

## Acceptance Criteria

- Durable decision/handoff records and file-authoritative ticket state survive SQLite rebuild.
- Codex runtime can execute in isolated worktrees with events, transcripts, checkpoints, and squash landing.
- T-1003 can run as the first Codex-assisted implementation ticket through Groundwork.
