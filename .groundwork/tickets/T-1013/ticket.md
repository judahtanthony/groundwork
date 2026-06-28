---
id: T-1013
kind: ticket
node_type: leaf
work_type: technical_design
title: Elevation-readiness suggestion queue
status: done
assignee: human.owner
requested_actor: null
priority: 0.6
labels: []
parent: T-1009
depends_on:
    - T-1012
created_at: "2026-06-21T20:08:43Z"
updated_at: "2026-06-26T00:57:28Z"
---

## Problem

Per ADR 0038: surface candidates ready for loosening for human review (the dashboard anticipates this). Suggest, never perform.

## Acceptance Criteria

- The system can suggest elevations for human review; it never self-elevates.
