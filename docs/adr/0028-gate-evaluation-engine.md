# ADR 0028: Gate Evaluation Engine

Status: Accepted
Implemented: Partial

> Partially superseded by ADR 0038: the current implementation still uses the
> reversibility short-circuit described here, while the accepted direction is to
> route irreversibility through policy as a highest-bar condition.

## Context

Several inputs decide whether a gated action (`execute`, `decompose`,
`land_to_main`, `replan`) may proceed: trust policy (`policies.md`, evaluated
top-down, first-match, stable rule IDs), reversibility ([ADR 0014](0014-reversibility-as-a-gate-input.md)),
risk class (`trust-and-approvals.md`), validation state (`validation.md`), and
actor authorization ([ADR 0023](0023-actors-work-types-and-policy-routing.md),
[ADR 0029](0029-actor-identity-model.md)). These are described across documents
but never composed into one algorithm, and two of them appear to conflict:
policy is *first-match ordered*, while actor matching is *prefix/set* based. Phase
2 needs one deterministic, auditable decision procedure.

## Decision

A single **gate evaluation engine** in `internal/policy` (with risk in
`internal/risk`, reversibility classification in the same area) that takes an
action context and returns a **`Decision` value with no side effects** (the
engine produces decisions, not mutations — the boundary rule in `overview.md`).
The decision records which rule ID fired, the risk class/score, the reversibility
verdict, and the required approver constraints, so the surface is fully
explainable.

When a gate must survive rebuild, the policy `Decision` is recorded in a durable
ticket-attached request/gate record and the approval row is only the live queue
projection (ADR 0051). The engine still produces side-effect-free decisions; the
coordinator decides how to persist and project them.

Composition order (later steps cannot loosen earlier floors):

1. **Classify scope.** From the action's changed files, commands, and external
   effects, compute reversibility and a 0–100 risk score mapped to a class
   (`low`/`medium`/`high`/`critical`). Score ranks and explains; it does not gate.
2. **Reversibility floor.** Irreversible ⇒ force `critical`, human-required, never
   auto-approved, regardless of score ([ADR 0014](0014-reversibility-as-a-gate-input.md)).
   This floor is non-overridable by any later step.
3. **Policy rules, first-match.** Evaluate trust rules top-down; the first
   matching rule decides (`policies.md`). Rule order is the spine and the fired
   rule ID is recorded.
4. **Validation gate.** For `land_to_main`, required validations for the changed
   files must pass (or carry an audited human override) independent of the trust
   rule ([ADR 0008 validation in validation.md], T-0703).

**Reconciling first-match with actor matching.** Rule *ordering* stays first-match;
actor matching is a **parameterized predicate evaluated within each rule**, not a
competing ordering scheme. A rule matches when all of its present conditions hold:
identity prefix (`actor_ids` matched as dotted-path prefixes per
[ADR 0029](0029-actor-identity-model.md)), `actor_types`, `roles` (set),
`work_types` (set), `action`, `files`, `risk_class[_at_most]`, and `reversible`.
Absent conditions are wildcards. So "first matching rule wins" and "actors match
by prefix/set" coexist: ordering selects the rule, the predicate decides whether
the rule applies.

**Precedence**, highest first: irreversible/critical floor → `require_human`
rules → `auto_approve` rules → `allow_claim` (claim/route authorization) →
**default deny**. A `requested_actor` on a node is only a routing hint into actor
selection; it never substitutes for an `allow_claim` match
([ADR 0023](0023-actors-work-types-and-policy-routing.md)).

## Consequences

`internal/policy` exposes one `Evaluate(ctx) Decision` used by the scheduler
(claim authorization + actor selection), the approval flow (gate type + required
approver), and landing (validation gate). The coordinator consults the engine
live for landing: `POST /tickets/:id/land` opens a `land_to_main` approval via
`Request`/`Evaluate`, so a node lands only by policy auto-approval or an approved
human gate. In M2 the changed-file `Scope` for `execute`/`land_to_main` is empty
(it arrives with the Phase 6 runtime's diff), so risk-scored auto-approval of
landing — including documentation auto-approval — is degenerate until then and
landing is effectively human-gated; the engine path is wired and exercised so
Phase 4 only supplies the diff. Autonomy levels (`policies.md`) select the
*default* for an action class but cannot drop below the reversibility floor;
elevation stays a human act ([ADR 0011](0011-progressive-planning-autonomy-via-sops.md)).
Chat-approval and reviewer-agent adapters are out of M2 (roadmap defers them to
"Phase 2 product features beyond v1"); when added, they must write the same
actor-aware durable request records and approval projections and cannot override
`require_human` — the engine is where that invariant is enforced. Risk *scoring*
refinement and earned/revocable autonomy remain Phase 5.
