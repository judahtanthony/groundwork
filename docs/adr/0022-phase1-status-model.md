# ADR 0022: Phase 1 Status Model Is Store-Validated; Gate Transitions Deferred

Status: Accepted
Implemented: Implemented

## Context

`docs/architecture/state-model.md` defines the status set (`backlog`, `todo`, `in_progress`, `blocked`, `review`, `rework`, `approved`, `landing`, `done`, `cancelled`) but not which transitions are legal or who authorizes them. Phase 1 has no coordinator or approvals; `gw ticket transition` is a manual human command.

## Decision

In Phase 1 the store validates that a target status is a member of the defined set and applies a minimal legal-transition map for manual use (for example `backlog ↔ todo`, `todo → in_progress`, `→ blocked`, `→ done`, `→ cancelled`). Gate-controlled transitions (`review → approved → landing`, `rework` cycles, escalation-driven `blocked`) are recognized as valid states, but their authorization — approvals, validation, and reversibility — is Phase 2. Every transition appends an audit event with actor and from/to. Rollup-derived parent state (`docs/architecture/work-tree.md`) is computed, never written by `transition`.

## Consequences

Phase 1 has a usable, auditable manual workflow without prejudging Phase 2 gating. The transition map lives in one place so the coordinator can later wrap it with policy.
