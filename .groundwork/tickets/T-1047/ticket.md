---
id: T-1047
kind: ticket
node_type: leaf
work_type: technical_implementation
title: Upgrade Dashboard + screens to in-place realtime (event-driven fragment updates)
status: backlog
assignee: null
requested_actor: null
priority: 0.55
labels:
    - web-ui
parent: T-1036
depends_on:
    - T-1037
    - T-1039
created_at: "2026-06-24T15:37:14Z"
updated_at: "2026-06-24T15:37:14Z"
---

## Problem

Replace full-page reload with targeted in-place updates driven by the event stream (KPI values, new timeline rows, cards moving lanes) per ADR 0042. No surface regresses below live data.

## Acceptance Criteria

_None recorded._
