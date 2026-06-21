---
id: T-0503
kind: ticket
node_type: leaf
work_type: technical_implementation
title: Stream runtime events to store and JSONL
status: backlog
assignee: null
requested_actor: null
priority: null
labels: []
parent: E-0006
depends_on:
    - T-0502
created_at: ""
updated_at: ""
---

## Problem

_No description recorded._

## Acceptance Criteria

- Run events persist in SQLite.
- events.ndjson is appended locally.
- Run records include actor_id and runtime/model metadata.
