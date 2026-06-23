---
id: T-1028
kind: ticket
node_type: leaf
work_type: technical_implementation
title: Add per-command flag help (Command.Flags/Long) and backfill leaf commands
status: done
assignee: null
requested_actor: null
priority: 0.75
labels:
    - cli-ux
parent: T-1022
depends_on:
    - T-1023
created_at: "2026-06-22T15:39:30Z"
updated_at: "2026-06-23T19:52:38Z"
---

## Problem

Extend the Command type so leaf commands can declare their flags/usage, and render them in printHelp instead of the hard-coded '[--json]'. Backfill real flag lists for gw ticket list/create/edit, gw board, and other leaf commands so -h is self-documenting.

## Acceptance Criteria

- Command type lets leaf commands declare flags; printHelp renders a Flags section (plus the universal --json) instead of the hardcoded [--json]
- Real flag lists backfilled for ticket list/create/edit/claim, context, next, and land; flagless commands stay terse
