---
id: T-1081
kind: ticket
node_type: leaf
work_type: technical_implementation
title: Gated root land_to_main merge and cleanup
status: todo
assignee: null
requested_actor: null
priority: 0.7
labels:
    - phase-5
parent: T-1075
depends_on:
    - T-1080
created_at: "2026-06-25T20:23:29Z"
updated_at: "2026-06-25T20:24:08Z"
---

## Problem

On approved root landing, gw merges the integration branch into main (--no-ff default, configurable) and deletes the branch (ADR 0058).

## Acceptance Criteria

- Approved root landing performs the gated merge to main and removes the integration branch.
