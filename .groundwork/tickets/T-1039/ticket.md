---
id: T-1039
kind: ticket
node_type: leaf
work_type: technical_implementation
title: Bootstrap embedded static SPA (Vite) served by gw server via go:embed
status: backlog
assignee: null
requested_actor: null
priority: 0.58
labels:
    - web-ui
parent: T-1036
depends_on:
    - T-1037
created_at: "2026-06-24T15:37:14Z"
updated_at: "2026-06-24T22:23:28Z"
---

## Problem

Full embedded static SPA bootstrap for ADR 0042. Deferred behind the urgent operator-unblock slice (T-1061) unless the executor determines the SPA shell is the smallest practical way to deliver that slice. No runtime Node for users; static assets remain embedded in the gw binary.

## Acceptance Criteria

_None recorded._
