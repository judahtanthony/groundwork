---
id: T-1024
kind: ticket
node_type: leaf
work_type: technical_implementation
title: Extract value-ordering into a shared eligible-ordering surface
status: backlog
assignee: null
requested_actor: null
priority: 0.9
labels:
    - cli-ux
parent: T-1022
depends_on:
    - T-1023
created_at: "2026-06-22T15:39:30Z"
updated_at: "2026-06-22T15:39:47Z"
---

## Problem

Move orderByValue/priorityPath/comparePath out of internal/scheduler so both the scheduler and the CLI read path compute the same ADR 0039 ordering from one source. Add e.g. db.ListEligibleOrdered() (todo + deps satisfied, value-ordered) and have the scheduler consume it. Behavior-preserving; covered by existing scheduler tests plus new store tests.

## Acceptance Criteria

_None recorded._
