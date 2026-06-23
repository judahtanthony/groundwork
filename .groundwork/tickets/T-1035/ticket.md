---
id: T-1035
kind: ticket
node_type: leaf
work_type: technical_implementation
title: Reconcile completed composites to done so they leave the eligible set
status: backlog
assignee: null
requested_actor: null
priority: 0.4
labels:
    - cli-ux
    - scheduler
parent: G-0001
depends_on: []
created_at: "2026-06-23T21:27:29Z"
updated_at: "2026-06-23T21:27:29Z"
---

## Problem

A composite whose children are all terminal+done keeps its stored status (e.g. todo), so it stays in ListEligible / gw next / gw ticket list --ready and pollutes the picker. Observed: T-1022 lingered as todo after its children completed and a store rebuild. ComputeRollup (internal/ticket/rollup.go) already derives done-ness from children, but nothing writes it back to the parent's stored status, and eligibility ignores it for composites. Fix: reconcile a composite to done when all children are terminal+done (on child landing and/or startup reconciliation), or exclude all-children-done composites from the eligible set. Surfaced closing the T-1022 epic.

## Acceptance Criteria

_None recorded._
