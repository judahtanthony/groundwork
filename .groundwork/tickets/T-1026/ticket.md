---
id: T-1026
kind: ticket
node_type: leaf
work_type: technical_implementation
title: Add guided gw ticket claim verb
status: backlog
assignee: null
requested_actor: null
priority: 0.8
labels:
    - cli-ux
parent: T-1022
depends_on:
    - T-1024
created_at: "2026-06-22T15:39:30Z"
updated_at: "2026-06-22T15:39:47Z"
---

## Problem

gw ticket claim <id> [--actor <id>]: verify the node is eligible, set assignee, transition todo->in_progress, and print the context brief and next-step hint in one guided step. Refuses to claim ineligible/blocked nodes with a clear message.

## Acceptance Criteria

_None recorded._
