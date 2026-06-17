# Trust And Approvals

Approvals are capability gates. They unlock specific actions and must be auditable.

## Risk Classes

- `low`: auto-approval allowed by policy.
- `medium`: human approval required in v1.
- `high`: human approval required in v1.
- `critical`: human approval required; no auto-approval.

## Risk Score

Risk is shown as a 0–100 score mapped onto the classes above:

- `low`: 0–33,
- `medium`: 34–66,
- `high`: 67–100.

Certain rule matches force `critical` regardless of score — for example destructive commands, non-reversible migrations, or production credential access. Policy gates key off the class, not the raw number; the score is for display and ranking.

## Reversibility

Reversibility is a first-class gate input alongside risk (see [ADR 0014](../adr/0014-reversibility-as-a-gate-input.md)). Each gated action is classified from its scope (changed files, commands, database/network/external effects):

- **Reversible** (revertible via git, no external side effects): eligible for aggressive autonomy as SOPs and validations mature.
- **Irreversible** (non-reversible migrations, external/production state, destructive commands, credential access): forced to `critical` — human-required, never auto-approved, regardless of score.

Risk ranks and explains; reversibility sets the floor. This unifies `execute`, `decompose`, and `land_to_main` under one principle and is what lets human contribution trend toward zero for reversible low-risk work without weakening safety on irreversible actions.

## Default Examples

Auto-approve:

- Documentation-only changes to `AGENTS.md`, `MEMORY.md`, and `.groundwork/**/*.md` when the diff is small and classified as internal guidance.
- Safe test commands such as `go test`, `npm test`, `pytest`, or `make test` when run inside the workspace without network.

Require human approval:

- Edits to `.env*` or files containing secrets.
- Destructive commands.
- Landing to `main` in v1.
- Decomposition proposals (`decompose`) in v1.
- Production deploys or production credential access.

## Progressive Autonomy

High-leverage agent actions are treated uniformly as gated capabilities: leaf `execute`, `land_to_main`, and `decompose`. Each has a risk class and an approval requirement (an autonomy level) that can be loosened over time.

Loosening is earned as task-type SOPs, updatable task-type context, and defined validations mature. As an action class becomes well-defined and reliably approved, its autonomy level can move from human-required toward policy/auto. This generalizes the landing gate (see [ADR 0006](../adr/0006-human-landing-gate-v1.md) and [ADR 0011](../adr/0011-progressive-planning-autonomy-via-sops.md)): planning decomposition becomes progressively autonomous the same way execution and landing do.

## Phase 2

Phase 2 should add chat approvals and reviewer-agent approvals. These must write the same approval records as CLI/web decisions and must not silently override human-required rules.

## Policy Learning

Groundwork may suggest loosening any gated action — including `decompose` — after repeated clean human approvals, and may suggest refinements to the SOPs and context that justify it. Trust elevation is a human act in v1: Groundwork must not install learned rules or raise an autonomy level without explicit approval.

