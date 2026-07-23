---
id: T-1045
kind: ticket
node_type: leaf
work_type: technical_implementation
title: Policies surface (trust rules, validation templates, suggestion queue)
status: done
assignee: null
requested_actor: null
priority: 0.4
labels:
    - web-ui
parent: T-1036
depends_on:
    - T-1040
    - T-1092
created_at: "2026-06-24T15:37:14Z"
updated_at: "2026-07-23T12:06:02Z"
---

## Problem

Ordered trust rules with stable ids, validation templates by file type, a rule editor, and the policy-learning suggestion queue.

## Acceptance Criteria

- Shows ordered trust rules with their stable ids
- Shows validation templates grouped by file type
- Provides a rule editor to view and edit trust rules via the gated API
- Surfaces the policy-learning suggestion queue
- Implemented in the embedded SPA under web/ using the T-1040 design system and the T-1092 app shell/navigation; the server-rendered internal/server/web templates are the ADR-0042 interim and must not be modified. New backend work is JSON API under internal/server, and new or changed API endpoints must be documented in docs/contracts/http-api.md (now in envelope scope).
