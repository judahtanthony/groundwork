---
id: T-1044
kind: ticket
node_type: leaf
work_type: technical_implementation
title: Run detail screen (gw run; lands with the Codex runtime)
status: done
assignee: null
requested_actor: null
priority: 0.5
labels:
    - web-ui
parent: T-1036
depends_on:
    - T-1040
    - T-1092
created_at: "2026-06-24T15:37:14Z"
updated_at: "2026-07-23T00:35:19Z"
---

## Problem

Live transcript, plan, changed files, validations, token/cost metrics, linked approval; pause/resume/cancel. Depends on real runtime output (E-0006).

## Acceptance Criteria

- Shows a run transcript WITH per-event messages, read from the durable per-run events.ndjson log (ADR 0027) via internal/server using proj.RunsDir(); plus the run plan and changed files
- readRunTranscript MUST degrade gracefully on log content: skip malformed JSON lines, tolerate a torn/partial trailing line (a live run mid-append), and handle oversized lines (raise the bufio scanner buffer AND skip any line still over the cap) — returning the successfully-parsed events. It must NEVER 500 on transcript content; only a genuine read error may error, and a missing log yields empty events. Replace any test that asserts a 500 on a malformed transcript with one asserting graceful skipping
- Shows validations, token/cost metrics, and the linked approval if any
- User can pause, resume, and cancel a run from the UI via the gated API
- Reaches parity with gw run (show, pause, resume, cancel)
- MUST NOT modify internal/scheduler or internal/store: read the transcript from events.ndjson within internal/server (prior run R-0010 was rejected for a scheduler change)
- Implemented in the embedded SPA under web/ using the T-1040 design system and the T-1092 app shell/navigation; the server-rendered internal/server/web templates are the ADR-0042 interim and must not be modified. New backend work is JSON API under internal/server, and new or changed API endpoints must be documented in docs/contracts/http-api.md (now in envelope scope).
