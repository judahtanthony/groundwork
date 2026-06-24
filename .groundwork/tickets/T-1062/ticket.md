---
id: T-1062
kind: ticket
node_type: null
work_type: technical_implementation
title: 'Operator UI: approvals inbox'
status: backlog
assignee: null
requested_actor: null
priority: 0.86
labels:
    - web-ui
    - operator-unblock
parent: T-1061
depends_on:
    - T-1063
created_at: "2026-06-24T22:22:35Z"
updated_at: "2026-06-24T22:22:35Z"
---

## Problem

Add an approvals inbox grouped by risk/type with approval detail, requesting actor, required actor/role constraints, ticket context, and current gate reason.

## Acceptance Criteria

- UI lists pending approvals and can open approval details.
- Approval detail shows ticket id, type, risk class/score, reversible flag, summary, actor constraints, and action payload when available.
