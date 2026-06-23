---
id: T-1030
kind: ticket
node_type: leaf
work_type: technical_implementation
title: Make gw ticket tree --json include priority (and parent/work_type)
status: done
assignee: null
requested_actor: null
priority: 0.6
labels:
    - cli-ux
parent: T-1022
depends_on:
    - T-1023
created_at: "2026-06-22T15:39:30Z"
updated_at: "2026-06-23T19:42:08Z"
---

## Problem

tree --json currently omits priority, parent, and work_type that show --json includes. Align the tree JSON node shape so callers don't have to fall back to show --json. Preserve the nested children structure.

## Acceptance Criteria

- gw ticket tree --json includes parent_id, work_type, and priority, matching show --json
- Nested children structure preserved
