---
id: T-1090
kind: ticket
node_type: leaf
work_type: technical_implementation
title: Route scheduler AI claims through envelope-aware authorization
status: todo
assignee: null
requested_actor: null
priority: 0.55
labels:
    - phase-6
    - bounded-autonomy
parent: E-0012
depends_on:
    - T-0501
created_at: "2026-06-28T22:00:00Z"
updated_at: "2026-06-28T22:00:00Z"
---

## Problem

`internal/scheduler/scheduler.go` selects actors with the trust-only
`policies.AuthorizeClaim` (empty scope, no envelope). ADR 0056's envelope-aware path
`Server.AuthorizeEnvelopedClaim` exists but is never used by the scheduler. Route AI
dispatch through it so trust AND the active ancestor envelope both gate the claim, and a
boundary crossing opens a human exception approval instead of silently failing. Humans
still bypass envelopes; no-envelope AI claims remain default-deny.

## Acceptance Criteria

- The scheduler authorizes AI claims via the envelope-aware path (trust AND envelope).
- A boundary crossing raises an exception approval and does not dispatch the run.
- No active envelope => AI claim denied (unchanged default-deny); humans unaffected.
- Tests cover allow-within-envelope, deny-no-envelope, and exception-on-crossing.
