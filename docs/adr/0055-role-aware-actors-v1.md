# ADR 0055: Role-Aware Actors v1 — Planner, Coding, Reviewer

Status: Pending Review
Implemented: Partial

Refines [ADR 0048](0048-role-aware-actors-and-local-identity.md) into the minimal v1 role
model Phase 5 needs. Builds on the existing actor registry (`.groundwork/actors.yaml`,
`internal/actor`) and the policy matcher, which already accepts `roles`, `actor_types`,
and capability inputs (`internal/policy`, `Match`).

## Context

Bounded autonomy needs to express *who does what* and *who approves what* without yet
introducing multiple humans. The policy schema can already match roles; what is missing is
a settled v1 set of AI roles and the rule that a reviewer agent informs but never replaces
the human gate.

## Decision

Define three v1 AI roles plus the human owner, and make role a first-class policy and
approval input.

### Roles

- **planner** — proposes decomposition, parent contracts, and approval envelopes
  (ADR 0054). It may *request* `decompose`/`approve_envelope`; it may not approve them.
- **coding** (implementer) — claims and executes scoped child leaves within an approved
  envelope (ADR 0056).
- **reviewer** — inspects diffs, validation, and completion summaries and records
  **findings**. It does **not** decide human-gated approvals. Reviewer findings are inputs
  to human/bulk review (ADR 0057), never an approval. This preserves the ADR 0028/0038
  human-gate invariant: "approvals unlock actions, not people," and require_human is never
  bypassed by an agent.
- **owner** (human) — satisfies every role in v1, so a single engineer runs the whole
  model today; later, roles can map to different humans without changing the workflow.

### Actor IDs and registry

Use the namespaced convention from ADR 0048 (`<type>.<role>.<name>`):

```yaml
actors:
  - id: human.owner
    type: human
    roles: [owner, planner, coding, reviewer]   # owner satisfies all roles in v1
  - id: ai.planner.codex
    type: ai_agent
    roles: [planner]
  - id: ai.coding.codex
    type: ai_agent
    roles: [coding]
    capabilities:
      work_types: [technical_implementation, test_implementation, documentation]
  - id: ai.reviewer.codex
    type: ai_agent
    roles: [reviewer]
```

Keep ADR 0048's separation: id = identity, roles = authority/approval semantics,
capabilities = what an actor can do, runtime permissions = what a process may access. The
role segment in the id is for routing/display only; authority comes from `roles`.

Until the Phase 6 runtime exists, the AI roles are the *identities Claude-Code-directed
work claims under* — a human operating in the coding role claims as `ai.coding.codex` (or
`human.owner`), and the records reflect which role acted.

### Required approval roles

Approval records already carry `required_actors`/`required_roles` and a deciding actor
(`internal/store/sqlite`, `Approval`). v1 makes policy populate the required role from the
rule that fired (work type, scope, risk, action) and records which actor satisfied it. For
v1, `human.owner` satisfies all required roles; the records are honest about *which* role
was required and *why* (the firing rule), so the model is ready for multiple humans later.

## Deferred

- Multi-human local identity (`.groundwork/local.yaml`, `~/.config`) and `gw mine` /
  `--mine` filtering (ADR 0048) — not needed for single-operator Phase 5.
- Reviewer-agent *approval* authority (earned, sampled) — a later ADR; v1 reviewer is
  advisory only.
- Per-agent runtime permission profiles — arrive with the Phase 6 runtime.

## Consequences

- `.groundwork/actors.yaml` gains the three AI role actors; policy rules can target
  `roles: [coding]` etc.; the claim gate (ADR 0056) reads the claiming actor's role.
- Reviewer findings become a recorded artifact consumed by bulk review (ADR 0057) without
  touching the approval-decision path.

## Open Questions

- One reviewer actor vs. specialized reviewers (security/test/docs)? v1: one `reviewer`;
  specialization is additive later.
- Should role be inferred from the actor id segment or only from `roles`? v1: only from
  `roles` (id segment is advisory), per ADR 0048.
