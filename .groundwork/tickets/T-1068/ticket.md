---
id: T-1068
kind: ticket
node_type: composite
work_type: technical_implementation
title: Implement approval envelopes for parent/root work
status: todo
assignee: null
requested_actor: null
priority: 0.55
labels:
    - phase-5
    - autonomy
parent: T-1066
depends_on: []
created_at: "2026-06-24T22:23:00Z"
updated_at: "2026-06-25T20:24:08Z"
---

## Problem

Specify and implement the minimum approved-envelope model for parent/root work: allowed actions, work types, actors, file/resource scopes, validation requirements, risk ceilings, and exception triggers.

## Acceptance Criteria

- Envelope records can authorize bounded decomposition/execution/review preparation for children.
- Unexpected scope, failed validation, risk above ceiling, and contract changes create human-visible exceptions.

## Design / Contract

_No contract recorded._

## Escalations

_No escalations._
