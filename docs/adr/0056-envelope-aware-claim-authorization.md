# ADR 0056: Envelope-Aware Claim Authorization

Status: Accepted
Implemented: Partial

Composes [ADR 0028](0028-gate-evaluation-engine.md) (gate engine),
[ADR 0038](0038-authority-as-loosenable-gate.md) (authority as loosenable policy),
[ADR 0054](0054-approval-envelopes-v1.md) (envelopes), [ADR 0055](0055-role-aware-actors-v1.md)
(roles), [ADR 0046](0046-resource-scope-ownership-and-conflict-policy.md) (scope), and
[ADR 0058](0058-integration-targets-and-landing-levels-v1.md) (`land_to_parent` git
mechanics). This is the rule that turns an approved envelope into actual bounded autonomy.

## Context

The gate engine already evaluates `claim`/`execute` against trust-policy `allow_claim`
rules with rich match inputs (`internal/policy`, `Match`: roles, work types, files, risk).
What it does not yet know is whether an action sits **inside an approved envelope**. Without
that, an AI claim is either blanket-allowed by a trust rule (too loose) or always
human-gated (the bottleneck Phase 5 removes).

## Decision

Make an active, matching envelope a **necessary additional condition** for an AI agent to
claim/execute/land-to-parent a node — composed as **AND** with trust policy. The envelope
never loosens trust policy; it authorizes work *within* a human-approved boundary.

### Composition

For an AI actor, the action is authorized only when **both** hold:

1. **Trust policy** `allow_claim` (or the relevant gate) returns `allow` for the action,
   actor role, work type, scope, and risk (existing ADR 0028 evaluation), **and**
2. An **active envelope** on an ancestor (ADR 0054) authorizes this action: the action is
   in `approved_actions`, the node's work type is in `allowed_work_types`, the planned
   scope is within the envelope `scope.allow` (and not `deny`), the actor's role is in
   `allowed_roles`, and the risk class is at or below `risk_ceiling`.

Default-deny is preserved (ADR 0037): **no active envelope ⇒ no AI claim** (identical to
today, where AI claims are not auto-authorized). Human (`owner`) claims do not require an
envelope — humans may always pick up their own work.

### New gate inputs

Extend `policy.Action` with the envelope facts the rule needs:

```go
type Action struct {
    // … existing fields …
    ActorRole    string  // resolved acting role (ADR 0055)
    EnvelopeID   string  // active ancestor envelope, "" if none
    WithinEnvelope bool  // action ∈ approved_actions ∧ scope ∧ role ∧ risk ∧ work_type
    PlannedScope []string
}
```

`WithinEnvelope` is computed by the coordinator from the ancestor envelope before
evaluation, so policy rules can match on it (e.g. an `allow_claim` rule with
`when: {within_envelope: true, actor_roles: [coding], risk_class_at_most: medium}`).

### Exceptions vs. denial

An AI action that matches no trust rule is **denied** (default-deny). An AI action that a
trust rule would allow but which **falls outside the envelope** — unexpected/`deny` files,
risk above ceiling, disallowed work type, or contract change — does **not** silently fail:
it raises an **exception approval** (`require_human`) linked to the parent envelope
(ADR 0054 step 5). This keeps the human in the loop exactly at the boundary crossing, not
for in-bounds work.

### Ordering with ADR 0038

The reversibility/authority composition of ADR 0038 still applies: irreversible actions,
`amend_policy`/`elevate_autonomy`, and root `land_to_main` remain at their high default
bar regardless of envelope. An envelope can authorize bounded in-scope child work; it
cannot authorize crossing those higher bars.

## Pre-runtime value

This is the lever that makes Phase 5 useful before the Codex runtime: a human approves one
envelope, and thereafter the coding role (a human directing Claude Code, or later an agent)
can claim and land child leaves to the parent target without a per-child approval — while
any boundary crossing still surfaces for review.

## Consequences

- The coordinator resolves the active ancestor envelope and computes `WithinEnvelope`
  before each `claim`/`execute`/`land_to_parent` gate evaluation.
- `allow_claim` rules can be written against `within_envelope` + role + scope + risk.
- Exception approvals become a distinct, parent-grouped queue in the inbox (Phase 4 UI
  extended) and CLI.

## v1 limitations (diff-dependent enforcement deferred to Phase 6)

Two parts of the envelope boundary are intentionally inert in v1 because they require a
real change set, which only the Phase 6 runtime produces:

- **File-scope on an empty change set is permissive.** `envelopeScopeAllows` returns true
  when no files are presented (`PlannedScope`/changed files empty). Pre-runtime there is no
  diff to check, so scope is best-effort: it constrains a claim *only* when a file set is
  supplied. Full allow/deny enforcement arrives with the runtime diff (ADR 0046). This is a
  deliberate v1 choice, not a gap to close before merge.
- **`scope.files.require_review` and the `escalation.*` triggers are recorded but not yet
  enforced.** They key off boundary-crossing facts derived from an actual diff
  (unexpected files, contract change, validation failure, risk above ceiling, public-API
  change). Until the runtime supplies those facts, only the static axes — approved action,
  work type, role, risk ceiling, and (when a file set is present) file scope — gate a claim;
  a crossing on a static axis still raises an exception. Wiring `require_review` and the
  escalation triggers is Phase 6 work alongside the diff.

## Open Questions

- Should `land_to_parent` reuse `land_to_main`'s gate type with a different target, or be a
  new action type? v1 proposes a distinct `land_to_parent` action so root landing stays
  unambiguously human-gated.
- How is "planned scope" supplied pre-runtime (no diff yet)? v1: from the node's declared
  scope/contract; actual-vs-planned comparison (ADR 0046) lands with the runtime diff.
