---
id: T-1024
kind: ticket
node_type: leaf
work_type: technical_implementation
title: Extract value-ordering into a shared eligible-ordering surface
status: done
assignee: null
requested_actor: null
priority: 0.9
labels:
    - cli-ux
parent: T-1022
depends_on:
    - T-1023
created_at: "2026-06-22T15:39:30Z"
updated_at: "2026-06-22T20:30:09Z"
---

## Problem

Move orderByValue/priorityPath/comparePath out of internal/scheduler so both the scheduler and the CLI read path compute the same ADR 0039 ordering from one source. Add e.g. db.ListEligibleOrdered() (todo + deps satisfied, value-ordered) and have the scheduler consume it. Behavior-preserving; covered by existing scheduler tests plus new store tests.

## Acceptance Criteria

- db.ListEligibleOrdered() returns the eligible set (todo + deps satisfied) value-ordered per ADR 0039 (priority path, then id)
- The scheduler consumes the shared ordering; value-ordering logic no longer lives only in internal/scheduler
- Behavior-preserving: existing scheduler/ordering tests pass, moved into the store package
