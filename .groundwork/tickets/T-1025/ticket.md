---
id: T-1025
kind: ticket
node_type: leaf
work_type: technical_implementation
title: Add gw ticket list --ready and --blocked views
status: done
assignee: null
requested_actor: null
priority: 0.85
labels:
    - cli-ux
parent: T-1022
depends_on:
    - T-1024
created_at: "2026-06-22T15:39:30Z"
updated_at: "2026-06-22T20:36:59Z"
---

## Problem

Add --ready (eligible set: todo + deps satisfied, value-ordered via the shared surface) and --blocked (todo nodes with unsatisfied deps, annotated with which deps are not done) to gw ticket list. --json parity. This is the read surface the eligibility engine never had.

## Acceptance Criteria

- gw ticket list --ready shows the eligible set (todo + deps satisfied) in ADR 0039 value order, via db.ListEligibleOrdered()
- gw ticket list --blocked shows todo nodes with unsatisfied deps, annotated with the blocking dep ids and statuses
- --status, --ready, --blocked are mutually exclusive; --json parity for all three
