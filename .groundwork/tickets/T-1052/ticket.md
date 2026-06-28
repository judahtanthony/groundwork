---
id: T-1052
kind: epic
node_type: null
work_type: technical_design
title: Implement durable async handoff and decision routing
status: in_progress
assignee: null
requested_actor: null
priority: 0.6
labels:
    - phase-6
    - async-agents
parent: T-1071
depends_on: []
created_at: "2026-06-24T17:43:12Z"
updated_at: "2026-06-24T22:30:34Z"
---

## Problem

Phase 6 runtime/backend execution substrate. Implement ADR 0051/0052 plus ADR 0053: ticket-attached durable decision records, rebuildable live queues, blocked-run handoff/resume packets, consequential decisions as work nodes, and file-authoritative durable ticket state. This is needed before autonomous background agents can safely stop on blockers and resume later.

## Acceptance Criteria

- Ticket decision sidecars are imported/exported deterministically.
- Pending approval/input/decision queues rebuild from durable ticket records.
- Blocked autonomous runs exit with durable handoff context and release capacity.
- Context assembly resumes later runs from durable ticket/run/canon state.
