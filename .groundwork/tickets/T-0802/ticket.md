---
id: T-0802
kind: ticket
node_type: leaf
work_type: technical_implementation
title: Implement ticket board and detail pages
status: todo
assignee: null
requested_actor: null
priority: 0.5
labels: []
parent: T-1036
depends_on:
    - T-0801
    - T-1040
created_at: "2026-06-22T12:20:48Z"
updated_at: "2026-06-24T22:23:28Z"
---

## Problem

Older ticket-board/detail-page work retained for continuity. The urgent operator-unblock portion is now split into T-1061 and its children so ticket visibility and approvals can ship before the full web UI/design-system path. This ticket now represents the broader board/detail page follow-through after the operator slice establishes the basic surfaces.

## Acceptance Criteria

- Board groups tickets by status.
- Ticket detail shows timeline, runs, approvals, and validation.
