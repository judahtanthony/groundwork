---
id: T-1102
kind: ticket
node_type: null
work_type: technical_implementation
title: Re-running a ticket leaves a stale completion summary from the prior run
status: backlog
assignee: null
requested_actor: null
priority: null
labels:
    - workflow
parent: T-1074
depends_on: []
created_at: "2026-07-22T13:23:51Z"
updated_at: "2026-07-22T13:23:51Z"
---

## Problem

The completion summary is stored per-ticket (.groundwork/tickets/<id>/completion.yaml, mirrored to SQLite) and is agent-authored/optional. When a ticket is re-run after a rejected run, and the new run does not record a summary, the PRIOR run's summary silently persists and misrepresents the new work. Observed on T-1092: run R-0006 (rejected, changed internal/server/*) wrote the summary; the re-run R-0007 changed 6 files under web/ but recorded no summary, so 'gw ticket summary show T-1092' still reports R-0006's internal/server file list. Impact: the human land/review gate and the ADR-0057 bulk review bundle can show a reviewer the wrong changed-file set for the work they are approving. Options: key completion records per-run rather than per-ticket, supersede/clear the ticket summary when a new run starts, or require the runtime to write a summary at run end (fall back to the run's own captured diff, which was correct here: the R-0007 diff event recorded changed_files=6).

## Acceptance Criteria

- A ticket re-run never reports a prior run's completion summary as the current one
- The land/review gate and review bundle show the changed-file set of the run actually being reviewed
- Regression test covers: run A records a summary, ticket is re-run as B, B's summary (or the absence of one) does not surface A's file list
