# Ticket Export Contract

Tickets are canonical as files under `.groundwork/tickets/` and projected into SQLite
for runtime queries, scheduling, and transactions. Mutating ticket state through `gw` or
the coordinator must update the durable export before reporting durable success, or write
a durable replay record that can complete the export after a crash (ADR 0053).

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

Each ticket directory may also contain `decisions.ndjson`, the durable sidecar for
ticket-attached input, approval, decision, rework, and recovery records. If a ticket is
`blocked`, in `review`, or ready for `rework` because of a pending request/proposal, the
record explaining that state must be exported in the sidecar before the worker exits.
See [decision-records.md](decision-records.md).

Direct edits to exported Markdown are not automatically applied to a running coordinator
in v1. `gw ticket import` rebuilds SQLite from files, and a future `gw sync` may
reconcile explicit file edits into the live projection.
