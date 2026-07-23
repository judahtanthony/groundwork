---
id: T-1042
kind: ticket
node_type: leaf
work_type: technical_implementation
title: Next/Ready surface (gw next, list --ready/--blocked)
status: done
assignee: null
requested_actor: null
priority: 0.68
labels:
    - web-ui
parent: T-1036
depends_on:
    - T-1040
    - T-1092
created_at: "2026-06-24T15:37:14Z"
updated_at: "2026-07-23T00:01:19Z"
---

## Problem

Surface the eligible set: top recommendation with brief, value-ordered ready list, blocked list annotated with blockers; claim from the UI.

## Acceptance Criteria

- Shows the top recommended next node with its context brief (parity with gw next)
- Lists the ready/eligible set in value (priority) order
- Lists blocked nodes, each annotated with its unmet blocker(s)
- User can claim a ready node from the UI via the gated API
- Implemented in the embedded SPA under web/ using the T-1040 design system and the T-1092 app shell/navigation; the server-rendered internal/server/web templates are the ADR-0042 interim and must not be modified. New backend work is JSON API under internal/server, and new or changed API endpoints must be documented in docs/contracts/http-api.md (now in envelope scope).
