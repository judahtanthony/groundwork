# ADR 0038: Authority As A Uniform Loosenable Gate

Status: Accepted
Implemented: Partial

Amends ADR 0011 (Progressive Planning Autonomy Via SOPs) and ADR 0014 (Reversibility As A
Gate Input).

## Context

ADR 0037 classifies "irreversible → human" and "elevation → human" as transitional
defaults, but the code and prior ADRs implement them as *structure*, not configuration:

- **Reversibility is a pre-policy floor.** The gate engine (`internal/policy/gate.go`,
  `Evaluate`) returns `require_human` for any irreversible action *before* consulting trust
  rules or autonomy levels. ADR 0014 states irreversible is "never auto-approved, regardless
  of score." No amount of earned trust, SOP maturity, or named constraint can pass it. This
  contradicts the near-term intent (ADR 0036) of agents making irreversible changes within
  constraints.

- **Authority-elevation is a human-only carve-out.** Raising an autonomy level or installing
  a learned rule has no action type and is reserved to humans by principle (ADR 0011,
  `trust-and-approvals.md` "Policy Learning"). The recursive "improvement" layer (ADR 0036
  Layer 3) is therefore structurally unreachable: the system can never be authorized to
  amend its own gates.

Both are the *same* shape of problem — a human wired in as structure rather than as a
high default on the gradient.

## Decision

Model both as ordinary gated actions evaluated by the one rule engine, so their human
requirement becomes a **default** expressed in policy rather than a wall in code.
Implementation is deferred to a future phase; this ADR fixes the direction and the
composition.

- **Reversibility becomes a highest-bar gate condition, not a short-circuit.** The
  reversibility verdict is still always computed and surfaced (an invariant, ADR 0037), but
  it flows *through* rule evaluation as a very-high-bar condition rather than returning
  before policy. An irreversible action is passable only by a sufficiently trusted actor
  under explicitly named constraints (e.g. sandbox, trust-tier ≥ N, required validations).
  Absent such a rule, it resolves to `require_human` — identical to today.

- **Authority-elevation becomes a first-class action type** (`amend_policy` /
  `elevate_autonomy`) subject to the same gate engine, default `require_human`. This makes
  delegating the "improvement" layer *expressible* without *enabling* it, and unifies it
  with `execute` / `decompose` / `land_to_main` / `review` instead of being a special case.

- **Default configuration preserves today's behavior exactly.** Shipped policy keeps
  irreversible → human and elevation → human. The only change is that conservatism is a
  *default rule*, not a structural guarantee — so it can be loosened deliberately, per
  action, within the invariant boundary of ADR 0037 (auditability, default-deny,
  reversibility-still-evaluated all remain).

## Consequences

When implemented, `gate.go`'s composition changes from "reversibility short-circuit →
rules" to "rules, with irreversibility as the highest bar," and the policy contract gains
the `amend_policy` / `elevate_autonomy` action types. This ADR implies but does not build
the supporting machinery: trust-tiers / earned autonomy wired through the currently unused
`AutonomyRequires{SOP, Validations}` (`internal/policy/policy.go`), and the
elevation-readiness "suggestion queue" the dashboard already anticipates
(`docs/architecture/dashboard.md`). Those are separate, later decisions. ADR 0011 and
ADR 0014 are amended: their human requirements are reclassified as loosenable policy
defaults, and irreversibility's *consequence* (not its evaluation) is what loosens.
