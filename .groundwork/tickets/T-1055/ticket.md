---
id: T-1055
kind: ticket
node_type: null
work_type: technical_implementation
title: Add blocked-run handoff outcomes and resume packets
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

Extend runtime outcomes and scheduler post-run handling so autonomous agents can exit blocked with durable handoff summaries, checkpoint refs, and resume packets for later runs.

## Acceptance Criteria

- Runtime result model distinguishes blocked, input_required, escalated, rework, completed, cancelled, and interrupted outcomes.
- Blocked outcomes move tickets to blocked with an explainable durable blocker, not review.
- New runs receive resume packets assembled from ticket/run/canon state, not a live model session.
