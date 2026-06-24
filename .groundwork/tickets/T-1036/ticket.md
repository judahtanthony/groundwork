---
id: T-1036
kind: epic
node_type: composite
work_type: null
title: 'Web UI: full-featured, realtime, CLI parity'
status: todo
assignee: null
requested_actor: null
priority: 0.7
labels:
    - web-ui
parent: G-0001
depends_on: []
created_at: "2026-06-24T15:37:14Z"
updated_at: "2026-06-24T15:37:14Z"
---

## Problem

Grow the web UI into the primary human interface at full parity with the CLI (ADR 0041), realtime-dynamic, top quality, per ADR 0042: contract-first JSON API + event stream (SSE now, WS-capable), a static-built SPA embedded in the gw binary via go:embed (single-binary, local-first preserved), progressive realtime (reload -> in-place fragment updates), all mutations routed through the same gates as the CLI. Sequenced below E-0006 (the Codex runtime is the active focus); the current server-rendered dashboard is the documented interim. Screen map from docs/architecture/dashboard.md.

## Acceptance Criteria

_None recorded._

## Design / Contract

_No contract recorded._

## Escalations

_No escalations._
