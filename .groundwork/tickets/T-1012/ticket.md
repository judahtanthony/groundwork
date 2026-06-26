---
id: T-1012
kind: ticket
node_type: leaf
work_type: technical_implementation
title: Wire trust-tiers / earned autonomy (AutonomyRequires)
status: done
assignee: human.owner
requested_actor: null
priority: 0.7
labels: []
parent: T-1009
depends_on:
    - T-1011
created_at: "2026-06-21T20:08:43Z"
updated_at: "2026-06-26T00:50:26Z"
---

## Problem

Per ADR 0038/0011: thread the currently-unused AutonomyRequires{SOP, Validations} so elevation can be earned within the ADR 0037 invariant boundary.

## Acceptance Criteria

- Autonomy levels gate on SOP + validation maturity.
