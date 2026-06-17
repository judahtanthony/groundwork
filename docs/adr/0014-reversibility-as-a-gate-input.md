# ADR 0014: Reversibility As A Gate Input

Status: Accepted

## Context

The trust model (ADR 0006, `trust-and-approvals.md`) gates actions with a hand-maintained list of human-required cases and an opaque 0–100 risk score. The variable that actually makes an action safe to auto-approve is whether it is **cheaply reversible**: a docs edit or an added test can be `git revert`-ed; a non-reversible migration, a production-credential touch, or a destructive command cannot. The approvals design already leans on this implicitly ("non-reversible migration +22, no tested rollback +22"), but it is not a first-class input.

## Decision

Make **reversibility a first-class, explicit input to every gate**, alongside risk class. Each gated action (`execute`, `decompose`, `land_to_main`, and individual tool actions) is classified by reversibility:

- **Reversible** (revertible via git, no external side effects): eligible for aggressive autonomy as SOPs and validations mature.
- **Irreversible** (non-reversible migrations, external/production state, destructive commands, credential access): forced to `critical`; human approval required and never auto-approved, regardless of score.

Reversibility is evaluated from the action's scope (changed files, commands, database/network/external effects), the same scope already shown on the approval surface. It composes with risk: risk ranks and explains; reversibility sets the floor. This unifies `execute`/`decompose`/`land_to_main` under one principle instead of a growing list, and it is what lets human contribution trend toward zero for reversible low-risk work without weakening safety on the irreversible cases.

## Consequences

`trust-and-approvals.md` and the policy contract gain a reversibility dimension; "irreversibility forces `critical`" becomes a rule rather than an example. Risk scoring (a Phase 5 refinement) and reversibility are separate axes: a low-risk but irreversible action still requires a human. Earned/revocable autonomy by outcome tracking is a separate, later decision (v2).
