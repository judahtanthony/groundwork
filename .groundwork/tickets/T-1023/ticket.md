---
id: T-1023
kind: ticket
node_type: leaf
work_type: documentation
title: Record the human-CLI operating model (ADR 0041)
status: done
assignee: null
requested_actor: null
priority: 0.95
labels:
    - cli-ux
parent: T-1022
depends_on: []
created_at: "2026-06-22T15:39:30Z"
updated_at: "2026-06-22T20:22:04Z"
---

## Problem

Write ADR 0041 capturing the human-CLI operating model and new command surface: gw next picker, gw ticket list --ready/--blocked, enriched gw status, guided gw ticket claim, reparent via gw ticket edit --parent, help-text framework, tree --json priority, land --preview. Reference/amend ADR 0011/0036/0039. This is the first ticket landed through gw's own gate for this epic (milestone toward T-1003).

## Acceptance Criteria

- ADR 0041 exists under docs/adr/ recording the human-CLI operating model and the new command surface
- ADR references/relates ADR 0011, 0016, 0028, 0034, 0036, 0039 as appropriate
- ADR is Accepted and the four dimensions (visibility, assignment, execution, review/landing) are covered
