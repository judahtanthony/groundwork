---
id: T-1027
kind: ticket
node_type: leaf
work_type: technical_implementation
title: Add gw next picker (top eligible + brief, --claim)
status: done
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
updated_at: "2026-06-22T21:09:18Z"
---

## Problem

gw next: show the single top eligible node (value-ordered) plus a compact context brief (ancestor spine, acceptance, deps) and the command to take it. --claim picks and claims the top node in one step (delegates to gw ticket claim). --json parity.

## Acceptance Criteria

- gw next shows the single top eligible node (value-ordered) with a compact context brief and the command to take it
- gw next --claim claims the top node in one step (delegates to the same claim path); --actor sets assignee
- Empty eligible set prints a clear message; --json parity for both plain and --claim
