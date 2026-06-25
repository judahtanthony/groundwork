---
id: T-1067
kind: ticket
node_type: leaf
work_type: technical_implementation
title: Define planner, coding, and reviewer AI actors
status: todo
assignee: null
requested_actor: null
priority: 0.61
labels:
    - phase-5
    - autonomy
parent: T-1066
depends_on: []
created_at: "2026-06-24T22:23:00Z"
updated_at: "2026-06-25T20:24:08Z"
---

## Problem

Define the v1 role-specific AI actor model for bounded autonomy: planner agents propose decomposition/envelopes, coding agents implement scoped child work, and reviewer agents inspect diffs/validation/summaries without approving human-gated actions.

## Acceptance Criteria

- Actor roles/capabilities for planner, coding, and reviewer agents are specified for .groundwork/actors.yaml and policy matching.
- Reviewer agents are explicitly prohibited from satisfying human-gated approvals in v1.
