---
id: T-1061
kind: epic
node_type: null
work_type: technical_design
title: Operator UI unblock slice
status: todo
assignee: null
requested_actor: null
priority: 0.95
labels:
    - web-ui
    - operator-unblock
parent: T-1036
depends_on: []
created_at: "2026-06-24T22:22:21Z"
updated_at: "2026-06-24T22:22:21Z"
---

## Problem

Urgent revised Phase 4 slice of the web UI. Build enough local operator UI to reduce the human approval bottleneck before full embedded SPA polish: ticket visibility, ready/blocked queues, approvals inbox, approval decisions, and land preview. Use the existing coordinator API/SSE and gate handlers; do not bypass CLI-equivalent policy behavior.

## Acceptance Criteria

- Human can see available tickets, ready work, blocked work, and pending approvals in the UI.
- Human can approve, reject, or clarify approval requests through the same coordinator approval service used by the CLI.
- Human can preview the diff for a landing approval before approving.
