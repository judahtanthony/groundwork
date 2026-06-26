---
id: T-1079
kind: ticket
node_type: leaf
work_type: technical_implementation
title: Root integration-branch lifecycle
status: done
assignee: human.owner
requested_actor: null
priority: 0.9
labels:
    - phase-5
parent: T-1075
depends_on:
    - T-1078
created_at: "2026-06-25T20:23:29Z"
updated_at: "2026-06-26T01:11:49Z"
---

## Problem

gw creates/tracks a per-root integration branch gw/root/<id>-<slug> on envelope approval; records it on the node (ADR 0058).

## Acceptance Criteria

- Approving a root envelope creates and records the integration branch; cleanup on successful root landing.
