# ADR 0024: Dependency Satisfaction And Rollup Terminality

Status: Accepted
Implemented: Implemented

## Context

A work node can legally reach two terminal states: `done` and `cancelled`
(`docs/architecture/state-model.md`). Two derived computations key off child
state and the original Phase 1 implementation treated the two terminals
inconsistently:

- **Dependency eligibility** (`todo` and all dependencies satisfied) counted only
  `done` as satisfied — correct, but it left a `cancelled` prerequisite blocking
  its dependents with no automatic recovery.
- **Rollups** folded `cancelled` into `done`, so a parent whose children were all
  cancelled rolled up to `done` — reporting abandoned work as completed.

These are two different questions about the same states, so they need two
deliberate answers rather than one shared rule.

## Decision

**Eligibility — satisfied means `done` only.** A `cancelled` prerequisite does
**not** satisfy a dependency. A cancellation signals a scope or requirement
change; per [ADR 0010](0010-dependencies-and-upward-revision.md) and
[ADR 0023](0023-actors-work-types-and-policy-routing.md) the correct response is
an explicit, human-gated re-plan that cancels or re-points the affected
dependents — not eligibility silently routing around the cancellation and
dispatching a dependent against the now-stale plan. In Phase 1 this resolution is
manual; the dependent remains ineligible until a human acts. The satisfaction
predicate lives in one place (`ticket.DependencyMet`) so the claim path, the
eligibility query, and the eligible-list all agree.

**Rollups — `done` and `cancelled` are both terminal.** A node is *settled* when
it is `done` or `cancelled`. A parent whose children are all settled rolls up to
`done` if any child is `done`, otherwise (all cancelled) to `cancelled`. Rollup
answers "is this settled?", which is distinct from eligibility's "was this
delivered?", so the two predicates differ by design.

## Consequences

`ticket.DependencyMet` is the single satisfaction rule; `ComputeRollup` is
terminal-aware. A `cancelled` node still blocks its dependents until a human
re-plans — intended conservative behavior in v1, and the seam where the
coordinator's escalation/re-plan flow surfaces and resolves such nodes. The
status enum is unchanged.
