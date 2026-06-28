---
id: T-1083
kind: ticket
node_type: leaf
work_type: technical_implementation
title: Envelope-aware gate composition and exception approvals
status: done
assignee: human.owner
requested_actor: null
priority: 0.8
labels:
    - phase-5
parent: T-1069
depends_on:
    - T-1082
created_at: "2026-06-25T20:23:29Z"
updated_at: "2026-06-26T01:26:57Z"
---

## Problem

Claim/execute/land_to_parent allowed only when trust policy AND an active envelope permit (AND); boundary crossings raise an exception approval (ADR 0056).

## Acceptance Criteria

- AI claim requires trust AND envelope; no envelope => denied; boundary crossing opens a parent-linked exception approval.
