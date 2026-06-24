---
id: T-1037
kind: ticket
node_type: leaf
work_type: technical_implementation
title: Harden the web-UI realtime API + event-stream contract (SSE now, WS-capable)
status: backlog
assignee: null
requested_actor: null
priority: 0.9
labels:
    - web-ui
parent: T-1036
depends_on: []
created_at: "2026-06-24T15:37:14Z"
updated_at: "2026-06-24T15:37:14Z"
---

## Problem

Make the coordinator JSON API + event stream a complete, stable UI contract per ADR 0042: cover every CLI capability (reads + mutations), a documented SSE event taxonomy keyed to UI regions, and a WebSocket upgrade path for bidirectional/high-frequency surfaces. Foundation the SPA builds on.

## Acceptance Criteria

_None recorded._
