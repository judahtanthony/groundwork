---
id: T-1063
kind: ticket
node_type: null
work_type: technical_implementation
title: 'Operator UI: tickets, ready, and blocked views'
status: todo
assignee: null
requested_actor: null
priority: 0.9
labels:
    - web-ui
    - operator-unblock
parent: T-1061
depends_on: []
created_at: "2026-06-24T22:22:35Z"
updated_at: "2026-06-24T22:22:35Z"
---

## Problem

Add the first operator-facing ticket views: ticket tree/list, value-ordered ready queue, and blocked queue with blocker explanations. Match CLI semantics for gw ticket tree, gw next/list --ready, and gw ticket list --blocked.

## Acceptance Criteria

- UI shows tickets with id, title, status, priority, work_type, node type, parent, and dependency/blocker state.
- Ready list matches value-ordered eligible work from the CLI/store.
- Blocked list shows unmet dependencies or durable blockers when available.
