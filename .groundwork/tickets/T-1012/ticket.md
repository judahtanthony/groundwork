---
id: T-1012
kind: ticket
node_type: leaf
work_type: technical_implementation
title: Wire trust-tiers / earned autonomy (AutonomyRequires)
status: backlog
assignee: null
requested_actor: null
priority: null
labels: []
parent: T-1009
depends_on: []
created_at: "2026-06-21T20:08:43Z"
updated_at: "2026-06-21T20:08:43Z"
---

## Problem

Per ADR 0038/0011: thread the currently-unused AutonomyRequires{SOP, Validations} so elevation can be earned within the ADR 0037 invariant boundary.

## Acceptance Criteria

- Autonomy levels gate on SOP + validation maturity.
