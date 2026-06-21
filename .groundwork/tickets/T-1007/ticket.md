---
id: T-1007
kind: ticket
node_type: leaf
work_type: technical_implementation
title: Add settable node priority field ([0,1], default 0)
status: todo
assignee: null
requested_actor: null
priority: null
labels: []
parent: T-1006
depends_on: []
created_at: "2026-06-21T20:08:26Z"
updated_at: "2026-06-21T20:08:26Z"
---

## Problem

Per ADR 0039: priority float [0,1] default 0 on nodes; settable via create/edit; carried through deterministic export (ADR 0020); inert until A2 consumes it.

## Acceptance Criteria

- priority is settable and round-trips through export/import.
- Default is 0; deterministic export preserved.
