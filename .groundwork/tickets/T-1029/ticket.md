---
id: T-1029
kind: ticket
node_type: leaf
work_type: technical_implementation
title: Enrich gw status with eligible/blocked/pending-approval counts
status: done
assignee: null
requested_actor: null
priority: 0.7
labels:
    - cli-ux
parent: T-1022
depends_on:
    - T-1024
created_at: "2026-06-22T15:39:30Z"
updated_at: "2026-06-23T19:40:47Z"
---

## Problem

Add eligible (ready) count, blocked count, and pending-approval count to gw status (text + --json), so 'what's the state?' answers what-can-I-do and what-needs-me at a glance.

## Acceptance Criteria

- gw status shows eligible (ready), blocked, and pending-approval counts in text and --json
- Counts match the eligibility engine (todo+deps satisfied = ready; todo+unmet dep = blocked)
