---
id: T-1058
kind: ticket
node_type: null
work_type: technical_implementation
title: Require completion and blocked-run handoff summaries
status: backlog
assignee: null
requested_actor: null
priority: 0.5
labels:
    - async-agents
parent: T-1052
depends_on: []
created_at: "2026-06-24T17:43:39Z"
updated_at: "2026-06-24T17:43:39Z"
---

## Problem

Implement ADR 0047 summary requirements for completed child work and blocked autonomous runs, including stale-summary invalidation after rework or parent integration changes.

## Acceptance Criteria

- Completion summaries are required before review/landing for runtime-produced results.
- Blocked runs write handoff summaries before lease release.
- Summary staleness is detected when result refs, parent contracts, dependencies, or rework decisions change.
