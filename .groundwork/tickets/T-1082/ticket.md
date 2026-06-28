---
id: T-1082
kind: ticket
node_type: leaf
work_type: technical_implementation
title: Envelope facts in policy.Action and WithinEnvelope computation
status: done
assignee: human.owner
requested_actor: null
priority: 0.9
labels:
    - phase-5
parent: T-1069
depends_on:
    - T-1081
created_at: "2026-06-25T20:23:29Z"
updated_at: "2026-06-26T01:20:37Z"
---

## Problem

Extend policy.Action with ActorRole/EnvelopeID/WithinEnvelope/PlannedScope; coordinator resolves the active ancestor envelope and computes WithinEnvelope (ADR 0056).

## Acceptance Criteria

- Gate inputs carry envelope facts; WithinEnvelope is computed from the ancestor envelope before evaluation.
