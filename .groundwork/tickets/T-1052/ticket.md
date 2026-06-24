---
id: T-1052
kind: epic
node_type: null
work_type: technical_design
title: Implement durable async handoff and decision routing
status: backlog
assignee: null
requested_actor: null
priority: 0.5
labels:
    - phase-4
    - async-agents
parent: G-0001
depends_on: []
created_at: "2026-06-24T17:43:12Z"
updated_at: "2026-06-24T17:43:12Z"
---

## Problem

Implement the architecture accepted in ADR 0051 and ADR 0052: ticket-attached durable decision records, rebuildable live queues, blocked-run handoff/resume packets, and consequential decisions as policy-routed work nodes.

## Acceptance Criteria

- Ticket decision sidecars are imported/exported deterministically.
- Pending approval/input/decision queues rebuild from durable ticket records.
- Blocked autonomous runs exit with durable handoff context and release capacity.
- Context assembly resumes later runs from durable ticket/run/canon state.
