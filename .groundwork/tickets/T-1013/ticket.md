---
id: T-1013
kind: ticket
node_type: leaf
work_type: technical_design
title: Elevation-readiness suggestion queue
status: backlog
assignee: null
requested_actor: null
priority: null
labels: []
parent: T-1009
depends_on: []
created_at: "2026-06-21T20:08:43Z"
updated_at: "2026-06-21T20:08:43Z"
---

## Problem

Per ADR 0038: surface candidates ready for loosening for human review (the dashboard anticipates this). Suggest, never perform.

## Acceptance Criteria

- The system can suggest elevations for human review; it never self-elevates.
