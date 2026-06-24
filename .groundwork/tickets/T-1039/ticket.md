---
id: T-1039
kind: ticket
node_type: leaf
work_type: technical_implementation
title: Bootstrap embedded static SPA (Vite) served by gw server via go:embed
status: backlog
assignee: null
requested_actor: null
priority: 0.8
labels:
    - web-ui
parent: T-1036
depends_on:
    - T-1037
created_at: "2026-06-24T15:37:14Z"
updated_at: "2026-06-24T15:37:14Z"
---

## Problem

Stand up a lean static-built SPA (React or Svelte + Vite) under web/, build to web/dist, embed via go:embed, serve same-origin from gw server (index at /, hashed assets at /assets, SPA fallback to index.html; API stays /api/v1). No runtime Node. Make target for the frontend build; decide committed-dist vs release-binary source-build (ADR 0042).

## Acceptance Criteria

_None recorded._
