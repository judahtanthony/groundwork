---
id: T-1039
kind: ticket
node_type: leaf
work_type: technical_implementation
title: Bootstrap embedded static SPA (Vite) served by gw server via go:embed
status: done
assignee: null
requested_actor: null
priority: 0.58
labels:
    - web-ui
parent: T-1036
depends_on: []
created_at: "2026-06-24T15:37:14Z"
updated_at: "2026-07-16T01:31:06Z"
---

## Problem

Full embedded static SPA bootstrap for ADR 0042. Deferred behind the urgent operator-unblock slice (T-1061) unless the executor determines the SPA shell is the smallest practical way to deliver that slice. No runtime Node for users; static assets remain embedded in the gw binary.

Design handoff (committed): scaffold the SPA to consume the ratified design reference at docs/design/groundwork-web-ui-design-system/ (tokens/, screens/, the .dc.html) and the IA at docs/design/web-ui-ia.md (ADR 0042 §IA).

## Acceptance Criteria

_None recorded._
