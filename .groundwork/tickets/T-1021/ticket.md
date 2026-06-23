---
id: T-1021
kind: ticket
node_type: null
work_type: technical_implementation
title: Support reparenting a node (gw ticket edit --parent or gw ticket move)
status: done
assignee: null
requested_actor: null
priority: null
labels: []
parent: G-0001
depends_on: []
created_at: "2026-06-22T12:56:58Z"
updated_at: "2026-06-23T20:00:27Z"
---

## Problem

Parent is fixed at creation: gw ticket edit has no --parent flag and there is no move/reparent command, so the work tree cannot be restructured without hand-editing exports (which desyncs the store and skips the audit trail). Add reparenting via the store with: cycle prevention (a node cannot be parented under its own descendant), parent-existence check, a recomputed rollup on both old and new parents, and a ticket.reparented audit event. Surfaced trying to move T-1019 to the root (G-0001).

## Acceptance Criteria

- A node's parent can be changed through gw (e.g. gw ticket edit --parent <id> or gw ticket move <id> --parent <id>).
- Reparenting under one's own descendant is rejected; an audit event is recorded; rollups on old and new parents update.
