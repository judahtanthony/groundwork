---
id: T-1091
kind: ticket
node_type: leaf
work_type: technical_implementation
title: Enforce envelope file-scope and escalation triggers using the run diff
status: todo
assignee: null
requested_actor: null
priority: 0.45
labels:
    - phase-6
    - bounded-autonomy
parent: E-0012
depends_on:
    - T-0507
    - T-1090
created_at: "2026-06-28T22:00:00Z"
updated_at: "2026-06-28T22:00:00Z"
---

## Problem

The envelope's `scope.files.require_review` and the five escalation triggers
(`on_unexpected_files`, `on_contract_change`, `on_validation_failure`,
`on_risk_above_ceiling`, `on_public_api_change`) are parsed and stored but never enforced,
and `envelopeScopeAllows` returns true on an empty file set. With the run diff available
(T-0507), enforce them at review/landing: a file outside the envelope's allow-list, a
touched `require_review`/contract/public-API path, a failed validation, or risk above the
ceiling raises a human exception approval rather than proceeding.

## Acceptance Criteria

- `envelopeScopeAllows` is enforced for real against the run's changed-file set.
- Each enabled escalation trigger that fires raises an exception approval at review/landing.
- The human-gated invariants are preserved (no trigger is bypassable when enabled).
- Tests cover each trigger firing and not-firing, and unexpected-file scope expansion.
