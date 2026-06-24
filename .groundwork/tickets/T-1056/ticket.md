---
id: T-1056
kind: ticket
node_type: null
work_type: technical_implementation
title: Extend context briefs with durable decisions and summaries
status: backlog
assignee: null
requested_actor: null
priority: 0.5
labels:
    - async-agents
parent: T-1052
depends_on: []
created_at: "2026-06-24T17:43:38Z"
updated_at: "2026-06-24T17:43:38Z"
---

## Problem

Include resolved/pending decision records, blockers, rework notes, dependency summaries, validation state, checkpoint refs, artifacts, and handoff/completion summaries in gw context.

## Acceptance Criteria

- gw ticket context exposes relevant durable decision records and blockers.
- Context assembly prefers summaries and canon over raw transcripts.
- Stale or missing summaries are surfaced as rework/recovery signals.
