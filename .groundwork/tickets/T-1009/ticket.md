---
id: T-1009
kind: epic
node_type: composite
work_type: technical_design
title: Authority as a uniform loosenable gate (ADR 0038)
status: done
assignee: null
requested_actor: null
priority: 0.6
labels: []
parent: T-1066
depends_on: []
created_at: "2026-06-21T20:08:43Z"
updated_at: "2026-06-26T00:57:28Z"
---

## Problem

Phase 5 authority-gate work. Implement ADR 0038 as part of bounded autonomy: reversibility and authority elevation are ordinary policy-gated actions, and human requirements are loosenable defaults only through explicit policy/envelope decisions. This reduces manual approval overhead without depending on Phase 6 runtime bring-up.

## Acceptance Criteria

- Conservatism is expressed as default policy rules within the ADR 0037 invariant boundary, not as code structure.

## Design / Contract

_No contract recorded._

## Escalations

_No escalations._
