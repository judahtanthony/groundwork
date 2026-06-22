---
id: T-1025
kind: ticket
node_type: leaf
work_type: technical_implementation
title: Add gw ticket list --ready and --blocked views
status: backlog
assignee: null
requested_actor: null
priority: 0.85
labels:
    - cli-ux
parent: T-1022
depends_on:
    - T-1024
created_at: "2026-06-22T15:39:30Z"
updated_at: "2026-06-22T15:39:47Z"
---

## Problem

Add --ready (eligible set: todo + deps satisfied, value-ordered via the shared surface) and --blocked (todo nodes with unsatisfied deps, annotated with which deps are not done) to gw ticket list. --json parity. This is the read surface the eligibility engine never had.

## Acceptance Criteria

_None recorded._
