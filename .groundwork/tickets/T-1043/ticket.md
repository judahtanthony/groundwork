---
id: T-1043
kind: ticket
node_type: leaf
work_type: technical_implementation
title: Approvals inbox (gw approval; land --preview)
status: done
assignee: null
requested_actor: null
priority: 0.48
labels:
    - web-ui
parent: T-1036
depends_on:
    - T-1040
    - T-1092
created_at: "2026-06-24T15:37:14Z"
updated_at: "2026-07-23T00:42:13Z"
---

## Problem

Full approvals inbox work for the CLI-parity web UI. The urgent Phase 4 approval visibility/decision work is split into T-1062/T-1064/T-1065 under T-1061; this ticket remains the broader polish/parity follow-through.

## Acceptance Criteria

- Lists all pending approvals (envelope, landing, decision) in one inbox
- User can approve, reject, or request clarification via the same coordinator approval service as the CLI
- For a landing approval, the diff preview is viewable before deciding (land --preview)
- Reaches parity with the gw approval CLI (list, show, decide)
- Implemented in the embedded SPA under web/ using the T-1040 design system and the T-1092 app shell/navigation; the server-rendered internal/server/web templates are the ADR-0042 interim and must not be modified. New backend work is JSON API under internal/server, and new or changed API endpoints must be documented in docs/contracts/http-api.md (now in envelope scope).
