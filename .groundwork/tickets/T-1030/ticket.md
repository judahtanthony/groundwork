---
id: T-1030
kind: ticket
node_type: leaf
work_type: technical_implementation
title: Make gw ticket tree --json include priority (and parent/work_type)
status: backlog
assignee: null
requested_actor: null
priority: 0.6
labels:
    - cli-ux
parent: T-1022
depends_on:
    - T-1023
created_at: "2026-06-22T15:39:30Z"
updated_at: "2026-06-22T15:39:47Z"
---

## Problem

tree --json currently omits priority, parent, and work_type that show --json includes. Align the tree JSON node shape so callers don't have to fall back to show --json. Preserve the nested children structure.

## Acceptance Criteria

_None recorded._
