---
id: T-1054
kind: ticket
node_type: null
work_type: technical_implementation
title: Rebuild live queues from durable decision records
status: backlog
assignee: null
requested_actor: null
priority: 0.5
labels:
    - async-agents
parent: T-1052
depends_on: []
created_at: "2026-06-24T17:43:26Z"
updated_at: "2026-06-24T17:43:26Z"
---

## Problem

Project pending durable approval/input/decision records into live SQLite queues on startup and detect inconsistent blocked/review/rework states.

## Acceptance Criteria

- Pending durable approval_requested records recreate approval queue rows with new runtime handles.
- Blocked or review tickets without durable explainers surface recovery_needed.
- DB purge/rebuild tests cover decompose, replan, land_to_main, input_required, and recovery_needed cases.
