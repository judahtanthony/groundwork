---
id: T-1046
kind: ticket
node_type: leaf
work_type: technical_implementation
title: Settings surface (paths, engine/sandbox, concurrency/lease, server, AGENTS.md sync, doctor)
status: done
assignee: null
requested_actor: null
priority: 0.38
labels:
    - web-ui
parent: T-1036
depends_on:
    - T-1040
    - T-1092
created_at: "2026-06-24T15:37:14Z"
updated_at: "2026-07-23T12:25:24Z"
---

## Problem

Repository/SQLite paths, agent engine + sandbox mode, concurrency + lease timing, server bind/port, AGENTS.md sync, gw doctor health checks.

## Acceptance Criteria

- Shows repository and SQLite paths, server bind/port, and concurrency + lease timing
- Shows agent engine and sandbox mode configuration
- Surfaces AGENTS.md sync status and action
- Runs and displays gw doctor health checks
- Implemented in the embedded SPA under web/ using the T-1040 design system and the T-1092 app shell/navigation; the server-rendered internal/server/web templates are the ADR-0042 interim and must not be modified. New backend work is JSON API under internal/server, and new or changed API endpoints must be documented in docs/contracts/http-api.md (now in envelope scope).
