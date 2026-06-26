# Phase 5 — Bounded Autonomy: implementation progress log

Autonomous execution of the Phase 5 leaf chain (ADRs 0054–0058) on
`phase-5-bounded-autonomy`. One entry per landed leaf; build/test green at each step.

- **T-1067** role-aware actors (ADR 0055) — added planner/coding/reviewer AI actors;
  `human.owner` holds all roles; `require_roles` is now a policy→approval input
  (gate `Decision.RequiredRoles` from the firing rule, recorded on the approval).
  Tests: gate required-roles + role-scoped match; approval records required role.
- **T-1010** reversibility as highest-bar condition (ADR 0038) — removed the
  pre-policy short-circuit in `gate.Evaluate`; irreversible actions now flow through
  rule evaluation and are passable only by an auto_approve rule that explicitly opts
  in (`when.reversible: false`), else `require_human`. Shipped defaults preserve
  prior behavior exactly. Tests: opt-in auto-approves; reversible-only/unqualified
  rules still gate.
- **T-1011** first-class `amend_policy` / `elevate_autonomy` action types (ADR 0038)
  — added the two approval types (Valid + human-gated set); they flow through the
  same gate engine and default to `require_human`. Authority elevation is now
  *expressible* without being *enabled*. Tests: types valid; default require_human.
- **T-1012** wire earned autonomy via `AutonomyRequires` (ADR 0038) — `autonomyOutcome`
  now honors a per-work-type elevation's prerequisites (named SOP present, required
  validations passed) read from new `Action.SatisfiedSOPs`/`PassedValidations`; unmet
  prerequisites fall back to require_human. Absent requirements stay back-compat.
  Tests: elevation gated when unmet, applies when met.
- **T-1013** elevation-readiness suggestion queue (ADR 0038) — new `policy_suggestions`
  table + store (list/get/set-status + `GenerateElevationSuggestions` scan: ≥3 clean
  done leaves of a work type with no failures → propose `execute@wt → auto`). New
  `gw policy suggestions [--scan] [--all]` / `promote` / `dismiss`. Promote emits the
  policy snippet for a human to apply — **never self-elevates** (amend_policy stays
  human-gated). Tests: scan creates/idempotent/dismiss; failure blocks. Completes the
  ADR 0038 authority-gate stream → 0038 Implemented: Partial.
- **T-1076** envelope record, store, file-authoritative sidecar (ADR 0054) — new
  `internal/envelope` package (full ADR 0054 shape + `Allows`/`AllowsRole`/
  `AllowsWorkType` + `envelope.yaml` sidecar read/write) and a SQLite mirror
  (`envelopes` table, migration 0005; upsert/get/get-active-for-node/set-status;
  one active envelope per node; status column authoritative on read). Tests:
  sidecar round-trip + missing; matchers; mirror CRUD + revoke.
