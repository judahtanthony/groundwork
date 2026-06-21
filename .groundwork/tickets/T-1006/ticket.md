---
id: T-1006
kind: epic
node_type: composite
work_type: technical_design
title: Value & prioritization substrate (Layer 1)
status: todo
assignee: null
requested_actor: null
priority: null
labels: []
parent: G-0001
depends_on: []
created_at: "2026-06-21T20:08:26Z"
updated_at: "2026-06-21T20:08:26Z"
---

## Problem

First slice of ADR 0036 Layer 1 per ADR 0039: a settable node priority and a value-ordered scheduler, replacing FIFO-by-id. Foundational; sequenced before M4.

## Acceptance Criteria

- Scheduler orders the eligible set by value (priority), not FIFO-by-id.
- priority is a live settable input, not an inert field.

## Design / Contract

_No contract recorded._

## Escalations

_No escalations._
