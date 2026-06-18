# Ticket Export Contract

Tickets are canonical in SQLite during runtime and exported to Markdown for durable committed state.

Example:

```md
---
id: T-0001
kind: ticket
node_type: leaf
work_type: technical_implementation
title: Implement SQLite migration runner
status: todo
assignee: null
requested_actor: ai.codex.default
priority: 2
labels: [store, sqlite]
parent: EPIC-store
depends_on: [T-0000]
created_at: 2026-06-17T10:00:00Z
updated_at: 2026-06-17T10:00:00Z
---

## Problem

Groundwork needs a migration runner before SQLite-backed features can be implemented.

## Acceptance Criteria

- Migrations apply in order.
- Re-running migrations is safe.
- Migration failures are surfaced clearly.
```

Composite nodes additionally carry a `## Design / Contract` section recording the schemas, interfaces, and requirements their children depend on, and an `## Escalations` section recording upward-revision events and re-plan decisions. `node_type` (`leaf` | `composite`) and `depends_on` reflect the triage outcome and dependency overlay. `work_type` is organization-defined operational metadata used by SOPs, policy, actor routing, and validation; it is not a status. `requested_actor` is an optional routing hint that policy must still authorize.

Direct edits to exported Markdown are not live state in v1. `gw ticket import` or `gw sync` may later reconcile explicit file edits.
