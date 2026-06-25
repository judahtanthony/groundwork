---
id: T-1069
kind: ticket
node_type: composite
work_type: technical_implementation
title: Add envelope-aware allow_claim policy
status: todo
assignee: null
requested_actor: null
priority: 0.45
labels:
    - phase-5
    - autonomy
parent: T-1066
depends_on: []
created_at: "2026-06-24T22:23:00Z"
updated_at: "2026-06-25T20:24:08Z"
---

## Problem

Extend claim authorization so AI agents can claim work only when both trust policy and an approved parent/root envelope allow the action, scope, actor role, risk class, and work type.

## Acceptance Criteria

- AI claims outside approved envelopes remain denied by default.
- Policy explanations identify the matching envelope and trust rule or the reason no claim was allowed.

## Design / Contract

_No contract recorded._

## Escalations

_No escalations._
