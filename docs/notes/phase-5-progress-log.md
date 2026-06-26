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
- **T-1077** envelope approval and lifecycle (ADR 0054) — new `approve_envelope`
  approval type (human-gated); `ProposeEnvelope` opens a pending approval carrying
  the draft, `activateEnvelope` (wired into the shared decision path) materializes
  it on approval (sidecar + mirror), and `RevokeEnvelope`/`SupersedeEnvelope` flip
  status in both. Child creation stays the decompose flow, composing within the
  envelope. Tests: propose→approve activates; revoke clears active.
- **T-1078** envelope CLI and operator-UI surface (ADR 0054) — `gw envelope
  list/show/revoke` (+ store `ListEnvelopes`); the approvals inbox renders the
  proposed boundary (actions/roles/work-types/risk/scope) for `approve_envelope`
  items so the human sees what they authorize. Exception-by-envelope grouping lands
  with the claim stream (exceptions are created there). Tests: inbox boundary shown.
- **T-1079** root integration-branch lifecycle (ADR 0058) — on envelope approval the
  root gets a recorded integration target (`integration_branches` table, migration
  0006): adopt the current feature branch, or create+checkout `gw/root/<id>-<slug>`
  from the default branch. New git helpers (HeadCommit/CreateAndCheckout/Checkout/
  MergeNoFF/DeleteBranch). No worktrees (Phase 6). Tests: slugify; branch created +
  checked out + recorded on approval.
- **T-1080** land_to_parent landing path (ADR 0058) — `LandToParent(childID)` resolves
  the nearest integration target (self/ancestors), checks it out, marks the child
  done, and commits its export + staged work there (reusing the ADR 0034 commit
  path) — distinct from land_to_main, no main merge. Endpoint
  POST /tickets/{id}/land-to-parent. Tests: child lands to integration branch (not
  main), HEAD advances, child done; errors without a target.
- **T-1081** gated root land_to_main merge and cleanup (ADR 0058) — `finishLanding`
  now, for a root with an open integration branch, merges it into the default
  branch (`--no-ff`), deletes the branch, and closes the record (`DefaultBranch`
  git helper). No-op for ordinary nodes. Completes the integration stream → ADR
  0058 Implemented: Partial. Test: end-to-end root land merges + cleans up.
- **T-1082** envelope facts in policy.Action + WithinEnvelope (ADR 0056) — added
  `ActorRole/EnvelopeID/WithinEnvelope/PlannedScope` to `policy.Action`; the
  coordinator resolves the active ancestor envelope (`activeAncestorEnvelope`) and
  computes `envelopeFacts`/`envelopeAuthorizes` (approved action ∧ work type ∧ role
  ∧ risk≤ceiling ∧ file scope allow/deny), reusing `policy.FilesMatch` + risk
  AtMost. Tests cover in-scope and each rejection axis + no-envelope.
- **T-1083** envelope-aware gate composition + exception approvals (ADR 0056) — added
  `Match.within_envelope`; `AuthorizeEnvelopedClaim` composes trust AND envelope for
  AI actors (humans bypass): no envelope ⇒ deny; trust+within ⇒ allow; trust-allowable
  but outside ⇒ a human `exception` approval (new approval type) linked to node+envelope;
  otherwise deny. Maps gate actions → envelope vocabulary. Completes the claim stream
  → ADR 0056 Implemented: Partial. Tests: human bypass, within→allow, crossing→exception,
  wrong-role/no-envelope→deny. (Live scheduler/runtime wiring is Phase 6.)
- **T-1084** completion-summary record (ADR 0047/0057) — new `internal/completion`
  package (Summary: outcome/changed/validation/decisions/assumptions/risks/canon +
  `completion.yaml` sidecar) and SQLite mirror (`completion_summaries`, migration
  0007). `gw ticket summary set/show`. The unit the bulk bundle aggregates. Tests:
  sidecar round-trip + mirror.
- **T-1085** review-bundle assembler + `gw review bundle` (ADR 0057) — `db.ReviewBundle`
  walks a subtree and deterministically aggregates per-leaf summaries, validations,
  and pending exceptions, with a recommendation (hold if unresolved exceptions,
  rework if a validation failed, else land). `gw review bundle <id> [--json]`. Tests:
  clean→land, failure→rework, exception→hold.
- **T-1086** web parent/root review screen (ADR 0057) — `GET /review/{id}` renders the
  bundle on the Phase 4 server-rendered chrome (recommendation badge, unresolved
  exceptions, per-child expandable summary/validation/exceptions); land_to_main inbox
  items link to it. Completes the bulk-review stream → ADR 0057 Implemented: Partial.
  Test: review page renders the bundle.

## Phase 5 complete
All 16 leaves landed on `phase-5-bounded-autonomy`; streams: role-aware actors,
authority gate (0038), envelopes (0054), integration/landing (0058), envelope-aware
claim (0056), bulk review (0057). Build/vet/full test suite green throughout. Phase 5
epic (T-1066) closure + merge to main await the human top-level review.
