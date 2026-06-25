---
id: T-1080
kind: ticket
node_type: leaf
work_type: technical_implementation
title: land_to_parent landing path
status: todo
assignee: null
requested_actor: null
priority: 0.8
labels:
    - phase-5
parent: T-1075
depends_on:
    - T-1079
created_at: "2026-06-25T20:23:29Z"
updated_at: "2026-06-25T20:24:08Z"
---

## Problem

Add the land_to_parent action: commit a child's work to its root integration target, generalizing ADR 0034's commit path (ADR 0058).

## Acceptance Criteria

- A child lands to the root integration branch (not main); land_to_parent is distinct from land_to_main.
