---
id: T-1065
kind: ticket
node_type: null
work_type: technical_implementation
title: 'Operator UI: land preview'
status: backlog
assignee: null
requested_actor: null
priority: 0.82
labels:
    - web-ui
    - operator-unblock
parent: T-1061
depends_on:
    - T-1062
created_at: "2026-06-24T22:22:35Z"
updated_at: "2026-06-24T22:22:35Z"
---

## Problem

Expose the same staged diff preview used by gw ticket land --preview so a human can inspect a landing approval from the UI before deciding.

## Acceptance Criteria

- UI can request and render the diff that would be committed for a landing approval.
- Preview handles empty staged diffs and command errors without mutating approval or ticket state.
