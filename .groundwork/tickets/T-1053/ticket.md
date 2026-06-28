---
id: T-1053
kind: ticket
node_type: null
work_type: technical_implementation
title: Add deterministic ticket decision sidecar import/export
status: done
assignee: null
requested_actor: null
priority: 0.75
labels:
    - async-agents
parent: T-1052
depends_on: []
created_at: "2026-06-24T17:43:26Z"
updated_at: "2026-06-24T17:43:26Z"
---

## Problem

Implement decisions.ndjson under ticket directories per docs/contracts/decision-records.md, including canonical encoding, import, export, and cold-rebuild byte-stability.

## Acceptance Criteria

- Ticket exports include decisions.ndjson when durable decision records exist.
- Import reconstructs decision records without changing canonical export bytes.
- Round-trip tests cover pending input, approval, decision, rework, and recovery records.
