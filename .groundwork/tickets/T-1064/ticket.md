---
id: T-1064
kind: ticket
node_type: null
work_type: technical_implementation
title: 'Operator UI: approval decisions'
status: backlog
assignee: null
requested_actor: null
priority: 0.84
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

Wire approve, reject, and clarify actions from the UI through the existing coordinator approval endpoints and ApprovalService. Preserve human-gated behavior; the UI must not self-approve or bypass policy.

## Acceptance Criteria

- Approve, reject, and clarify actions call the existing HTTP approval decision endpoints.
- Decision errors are displayed clearly and leave approval state unchanged.
- Approving land_to_main continues to perform the existing validated landing flow.
