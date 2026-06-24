---
id: T-1060
kind: epic
node_type: null
work_type: technical_design
title: 'Phase 4: Operator UI'
status: todo
assignee: null
requested_actor: null
priority: 0.98
labels:
    - phase-4
    - revised-plan
parent: G-0001
depends_on: []
created_at: "2026-06-24T22:22:04Z"
updated_at: "2026-06-24T22:33:15Z"
---

## Problem

Phase 4 focuses only on the local operator UI needed to reduce the human approval bottleneck before deeper autonomy or background runtime work: ticket visibility, ready/blocked queues, approvals inbox, approval decisions, and landing diff preview. It uses existing coordinator APIs and gates; it does not include durable async handoff or the Codex runtime.

## Acceptance Criteria

- Operator UI makes ready work, blocked work, and pending approvals visible and actionable.
- Human can approve, reject, or clarify approval requests through the existing coordinator approval service.
- Human can preview landing diffs before approving land_to_main.
