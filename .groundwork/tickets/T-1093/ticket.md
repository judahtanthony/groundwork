---
id: T-1093
kind: ticket
node_type: leaf
work_type: technical_implementation
title: 'Web UI: full CLI-parity ticket CRUD (view, browse, create, edit, transition, link)'
status: done
assignee: null
requested_actor: null
priority: 0.69
labels:
    - web-ui
parent: T-1036
depends_on:
    - T-1092
created_at: "2026-07-05T17:56:08Z"
updated_at: "2026-07-22T23:29:24Z"
---

## Problem

Bring the web UI to full CLI parity (ADR 0041/0042): today it is thin server-rendered read/approve/land pages. Add full ticket CRUD and the mutation surface on top of the DAG navigation from T-1092. Absorbs the mutation scope from the retired T-1041: create, edit, claim, triage, link, reparent, decompose, and escalate — all via the gated coordinator API (never bypassing policy/approval gates). Sequences after the DAG navigation redesign (T-1092).

## Acceptance Criteria

- Any read or mutation available in the gw CLI is available in the web UI
- User can create, view, browse, edit, transition, and link tickets from the web UI
- Node mutations claim, triage, reparent, decompose, and escalate are available via the gated API
- Implemented in the embedded SPA under web/ using the T-1040 design system and the T-1092 node view; the server-rendered internal/server/web templates are the ADR-0042 interim and must not be modified (new backend work is JSON API under internal/server only)
