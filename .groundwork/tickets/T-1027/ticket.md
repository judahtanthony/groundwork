---
id: T-1027
kind: ticket
node_type: leaf
work_type: technical_implementation
title: Add gw next picker (top eligible + brief, --claim)
status: backlog
assignee: null
requested_actor: null
priority: 0.78
labels:
    - cli-ux
parent: T-1022
depends_on:
    - T-1024
    - T-1026
created_at: "2026-06-22T15:39:30Z"
updated_at: "2026-06-22T15:39:47Z"
---

## Problem

gw next: show the single top eligible node (value-ordered) plus a compact context brief (ancestor spine, acceptance, deps) and the command to take it. --claim picks and claims the top node in one step (delegates to gw ticket claim). --json parity.

## Acceptance Criteria

_None recorded._
