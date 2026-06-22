---
id: T-0503
kind: ticket
node_type: leaf
work_type: technical_implementation
title: Stream runtime events to store and JSONL
status: todo
assignee: null
requested_actor: null
priority: null
labels: []
parent: E-0006
depends_on:
    - T-0502
created_at: "2026-06-22T12:20:48Z"
updated_at: "2026-06-22T15:06:06Z"
---

## Problem

_No description recorded._

## Acceptance Criteria

- Run events persist in SQLite.
- events.ndjson is appended locally.
- Run records include actor_id and runtime/model metadata.
