---
id: T-1022
kind: epic
node_type: composite
work_type: null
title: Human CLI workflow & observability
status: todo
assignee: null
requested_actor: null
priority: 0.95
labels:
    - cli-ux
    - observability
parent: G-0001
depends_on: []
created_at: "2026-06-22T15:39:08Z"
updated_at: "2026-06-22T15:39:47Z"
---

## Problem

Make the human CLI experience excellent before the web dashboard (E-0009) or AI runtime (E-0006). Surface the existing eligibility/priority engine to humans (what's ready, what's next, what's blocked, what needs approval), add guided assignment/execution/landing ergonomics, and fix help-text and --json consistency. Records the operating model in ADR 0041. Prioritized above E-0009 (0.9) since the CLI is how a human drives the whole loop today (ADR 0040). Re-triages T-1020 (cold-start) and T-1021 (reparent) into this epic once gw ticket edit --parent lands.

## Acceptance Criteria

_None recorded._

## Design / Contract

_No contract recorded._

## Escalations

_No escalations._
