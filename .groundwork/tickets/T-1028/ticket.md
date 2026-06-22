---
id: T-1028
kind: ticket
node_type: leaf
work_type: technical_implementation
title: Add per-command flag help (Command.Flags/Long) and backfill leaf commands
status: backlog
assignee: null
requested_actor: null
priority: 0.75
labels:
    - cli-ux
parent: T-1022
depends_on:
    - T-1023
created_at: "2026-06-22T15:39:30Z"
updated_at: "2026-06-22T15:39:47Z"
---

## Problem

Extend the Command type so leaf commands can declare their flags/usage, and render them in printHelp instead of the hard-coded '[--json]'. Backfill real flag lists for gw ticket list/create/edit, gw board, and other leaf commands so -h is self-documenting.

## Acceptance Criteria

_None recorded._
