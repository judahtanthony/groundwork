---
id: T-1057
kind: ticket
node_type: null
work_type: technical_implementation
title: Add consequential decision ticket routing path
status: todo
assignee: null
requested_actor: null
priority: 0.5
labels:
    - async-agents
parent: T-1052
depends_on:
    - T-1053
created_at: "2026-06-24T17:43:38Z"
updated_at: "2026-06-24T17:43:38Z"
---

## Problem

Support agents creating dependent decision tickets for consequential blockers, using normal work_type policy routing and dependency edges instead of a separate decision subsystem.

## Acceptance Criteria

- Agents can create a decision node and dependency edge before exiting blocked.
- Decision-oriented work_types route through existing SOP/policy/actor selection.
- Small clarifications remain local input requests rather than tickets.
