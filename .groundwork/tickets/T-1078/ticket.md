---
id: T-1078
kind: ticket
node_type: leaf
work_type: technical_implementation
title: Envelope CLI and operator-UI surface
status: done
assignee: human.owner
requested_actor: null
priority: 0.7
labels:
    - phase-5
parent: T-1068
depends_on:
    - T-1077
created_at: "2026-06-25T20:23:29Z"
updated_at: "2026-06-26T01:07:45Z"
---

## Problem

Expose envelopes via CLI (gw envelope show/approve/revoke) and the operator web UI; group exception approvals under their parent envelope (ADR 0054).

## Acceptance Criteria

- CLI and web show an envelope's boundary and status; exception approvals group by parent envelope.
