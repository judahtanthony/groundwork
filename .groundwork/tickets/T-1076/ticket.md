---
id: T-1076
kind: ticket
node_type: leaf
work_type: technical_implementation
title: Envelope record, store, and file-authoritative sidecar
status: done
assignee: human.owner
requested_actor: null
priority: 0.9
labels:
    - phase-5
parent: T-1068
depends_on:
    - T-1013
created_at: "2026-06-25T20:23:29Z"
updated_at: "2026-06-26T01:00:40Z"
---

## Problem

Add the approval_envelope record (ADR 0054): Go type, SQLite table, and the .groundwork/tickets/<id>/envelope.yaml sidecar (file-authoritative, SQLite mirror).

## Acceptance Criteria

- Envelope persists to SQLite and a committed envelope.yaml sidecar; survives a store rebuild from exports.
- Envelope shape matches ADR 0054 (actions, planning limits, file-glob scope, validation, risk ceiling, roles, escalation).
