---
id: T-1093
kind: ticket
node_type: null
work_type: technical_implementation
title: 'Web UI: full CLI-parity ticket CRUD (view, browse, create, edit, transition, link)'
status: backlog
assignee: null
requested_actor: null
priority: null
labels: []
parent: T-1036
depends_on: []
created_at: "2026-07-05T17:56:08Z"
updated_at: "2026-07-05T17:56:08Z"
---

## Problem

Today's web UI is thin server-rendered read/approve/land pages. Bring it to CLI parity (ADR 0041/0042): full ticket CRUD and navigation. Re-scopes the existing screen leaves under T-1036; sequence after the DAG navigation redesign.

## Acceptance Criteria

- Any read/mutation available in the gw CLI is available in the web UI
- User can create, view, browse, edit, transition, and link tickets from the web UI
