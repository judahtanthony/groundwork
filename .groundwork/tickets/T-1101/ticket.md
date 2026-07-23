---
id: T-1101
kind: ticket
node_type: null
work_type: technical_implementation
title: 'Retire rework status: route review->todo with a rework_requested decision record (ADR 0060)'
status: backlog
assignee: null
requested_actor: null
priority: null
labels:
    - workflow
parent: T-1074
depends_on: []
created_at: "2026-07-22T13:11:47Z"
updated_at: "2026-07-22T13:32:43Z"
---

## Problem

Implement ADR 0060. Retire the dead-end rework status and make review rejection usable end-to-end by REUSING the existing decision channel (no new sidecar). Today rework is unreachable for re-dispatch (dispatch is todo-only via txClaim; no rework->todo edge) and no code emits/consumes the already-defined rework_requested decision event (internal/decision EventReworkRequested). Touches internal/ticket (drop rework from status set + transition map; add guarded review->todo), internal/store/sqlite (atomic transition + decisions.ndjson write), internal/resume (read latest unaddressed rework_requested into NextAction/HandoffSummary), internal/cli (gw review reject verb), internal/server (review API/UI), the coder SOPs, and docs/contracts/decision-records.md + docs/architecture/state-model.md. OUTSIDE the T-1036 web envelope; platform/workflow work.

## Acceptance Criteria

- rework is removed from the status set; review->rework and rework->* edges are replaced by a single guarded review->todo edge (state-model.md updated)
- Rejection is recorded on the existing .groundwork/tickets/<id>/decisions.ndjson as a rework_requested event using the existing schema (statement carries the correction); no new sidecar/encoding/reader is introduced
- gw review reject <id> --note ... requires a non-empty note and writes the rework_requested record in the same transaction as review->todo
- resume.Assemble sets NextAction/HandoffSummary from the latest unaddressed rework_requested statement; a record is considered addressed once a later run on the ticket reaches review
- A review.rejected audit event is emitted; todo remains the single dispatchable status (eligibility/txClaim/rollups unchanged)
- The technical_implementation and test_implementation SOPs are updated so a re-dispatched agent checks for an unaddressed rework_requested record, treats the node as a continuation, honors scope/surface-narrowing corrections, and states how feedback was addressed
- docs/contracts/decision-records.md rebuild semantics are updated: rework_requested explains why a ticket was returned to todo, and an unaddressed record marks work awaiting correction
