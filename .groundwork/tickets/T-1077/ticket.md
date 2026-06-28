---
id: T-1077
kind: ticket
node_type: leaf
work_type: technical_implementation
title: Envelope approval and lifecycle
status: done
assignee: human.owner
requested_actor: null
priority: 0.8
labels:
    - phase-5
parent: T-1068
depends_on:
    - T-1076
created_at: "2026-06-25T20:23:29Z"
updated_at: "2026-06-26T01:04:33Z"
---

## Problem

Propose/approve an envelope via the approval service; on approval create children in backlog; support revoke and supersede (ADR 0054).

## Acceptance Criteria

- A human can approve an envelope for a composite/root; children are created in backlog.
- Revoke flips status and blocks new claims/landings; supersede links the replacement.
