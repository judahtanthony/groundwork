---
id: T-1011
kind: ticket
node_type: leaf
work_type: technical_implementation
title: First-class amend_policy / elevate_autonomy action types
status: backlog
assignee: null
requested_actor: null
priority: null
labels: []
parent: T-1009
depends_on: []
created_at: "2026-06-21T20:08:43Z"
updated_at: "2026-06-21T20:08:43Z"
---

## Problem

Per ADR 0038: authority-elevation becomes gated action types (default require_human), unified with execute/decompose/land_to_main/review.

## Acceptance Criteria

- amend_policy and elevate_autonomy are evaluated by the one gate engine, default require_human.
