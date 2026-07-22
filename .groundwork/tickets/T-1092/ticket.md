---
id: T-1092
kind: ticket
node_type: leaf
work_type: technical_implementation
title: 'Web UI: DAG-oriented navigation redesign (root nodes as primary handle)'
status: done
assignee: null
requested_actor: null
priority: 0.7
labels:
    - web-ui
parent: T-1036
depends_on:
    - T-1040
created_at: "2026-07-05T17:56:08Z"
updated_at: "2026-07-22T13:29:08Z"
---

## Problem

Implement the DAG-oriented navigation and node view — the primary Tickets screen. The web UI was designed before the DAG work-node model (ADR 0044/0045); root nodes (parentless, typically human-created) are the primary handle and users drill up and down the tree, while most branch/leaf nodes are agent-created. Design RATIFIED: docs/design/web-ui-ia.md + ADR 0042 §IA; hero comp docs/design/groundwork-web-ui-design-system/screens/nodeview.png (up·here·down spine, envelope-coverage provenance boundary, per-child breakdown, drag-to-prioritize, settled toggle). Implement against that reference. Absorbs the tickets-screen navigation and node-view scope from the retired T-1041: tree/outline with breadcrumbs, leaf/composite indicators, dependency edges, and status rollup; a node-detail (here) panel showing problem, acceptance, validations, runs, diff, and timeline. Mutation/CRUD parity is the sibling ticket T-1093 and sequences after this. SURFACE (required): implement in the embedded Vite SPA under web/ (served at /app/), consuming the T-1040 design-system components in web/src/design-system/. The server-rendered internal/server/web/* Go templates (dashboard, /tickets, /approvals) are the ADR-0042 INTERIM surface and MUST NOT be modified or targeted. New backend work is allowed only as JSON API endpoints under internal/server (e.g. /api/v1/...). Prior run R-0006 was rejected for building on the server-rendered /tickets page instead of the SPA.

## Acceptance Criteria

- Root nodes (no parent) are the top-level handle in the UI
- User can drill down parent to child and back up the DAG, with breadcrumbs
- Design accounts for human-created roots and agent-created branch/leaf nodes
- Nodes show leaf/composite indicators, dependency edges, and rollup status for composites
- The focused node detail panel shows problem, acceptance, validations, runs, diff, and timeline
- Implemented in the embedded SPA under web/ using the T-1040 design system; no changes to server-rendered internal/server/web templates (new backend work is JSON API under internal/server only)
