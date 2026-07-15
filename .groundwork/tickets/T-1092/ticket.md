---
id: T-1092
kind: ticket
node_type: null
work_type: technical_design
title: 'Web UI: DAG-oriented navigation redesign (root nodes as primary handle)'
status: backlog
assignee: null
requested_actor: null
priority: null
labels: []
parent: T-1036
depends_on: []
created_at: "2026-07-05T17:56:08Z"
updated_at: "2026-07-15T01:18:45Z"
---

## Problem

The web UI was designed before the DAG work-node model (ADR 0044/0045). Rethink navigation so root nodes (parentless, typically human-created) are the primary handle and users drill up/down the tree; most branch/leaf nodes will be agent-created. Precedes/reshapes the screen leaves (T-1041 tickets screen, etc.).

Design RATIFIED: docs/design/web-ui-ia.md + ADR 0042 §IA; hero comp docs/design/groundwork-web-ui-design-system/screens/nodeview.png (up·here·down spine, envelope-coverage provenance boundary, per-child breakdown, drag-to-prioritize, settled toggle). Implement against that reference.

## Acceptance Criteria

- Root nodes (no parent) are the top-level handle in the UI
- User can drill down parent->child and back up the DAG
- Design accounts for human-created roots and agent-created branch/leaf nodes
