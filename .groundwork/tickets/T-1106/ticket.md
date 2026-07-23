---
id: T-1106
kind: ticket
node_type: null
work_type: technical_implementation
title: Support un-parenting a node to a root; fix reparent API doc (empty parent claim)
status: backlog
assignee: null
requested_actor: null
priority: null
labels:
    - workflow
parent: T-1074
depends_on: []
created_at: "2026-07-23T13:44:47Z"
updated_at: "2026-07-23T13:44:47Z"
---

## Problem

Two related gaps found closing T-1036: (1) store Reparent (internal/store/sqlite/reparent.go:26) rejects an empty parent with 'new parent id is required', so an existing node cannot be un-parented into a standalone root — there is no command to make a node a root. Add support (CLI + POST /reparent) for reparenting to root (parent_id = NULL), guarding rollup/cycle invariants. (2) The landed docs/contracts/http-api.md (from T-1093) states 'an empty parent makes it a root', but handleTicketReparent calls the same Reparent that rejects empty — a doc/impl mismatch. Fix the doc or implement the behavior.

## Acceptance Criteria

- A node can be reparented to a root (no parent) via CLI and API, or the http-api.md reparent doc is corrected to match the reject-empty behavior
