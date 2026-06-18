# Actors

Groundwork treats humans and AI systems as **actors**. An actor is any local identity that can claim work, perform a run, review output, approve a gated action, or be recorded in audit history.

The v1 implementation stays local and single-user. It does not need accounts, authentication, or a team permission service. It needs a durable actor registry that lets policy and audit answer who or what was allowed to act, what capabilities were expected, and what configuration was used.

## Actor Types

Default actor types:

- `human`: a person operating the CLI or dashboard.
- `ai_agent`: an AI runtime that can execute or plan work.
- `ai_judge`: an AI reviewer that can judge work but normally cannot edit.
- `tool`: a non-agent automation such as validation or deployment.

Actor types are policy inputs, not hard security boundaries by themselves. Policy decides whether a specific actor, type, role, or capability may claim, review, approve, edit, or land.

## Actor Registry

Actors are defined in `.groundwork/actors.yaml` and committed by default. The scaffolded MVP registry should contain:

- `human.owner`: the single local owner, able to approve and review all v1 actions.
- `ai.codex.default`: the default Codex actor for planning, implementation, tests, and documentation.

The registry can later grow to include domain owners, security reviewers, product/design roles, documentation judges, task-specific AI agents, and actors backed by other runtimes or LLM providers.

## Capabilities

Capabilities describe routing and authorization inputs:

- work types an actor may claim,
- review or approval authority,
- roles such as `owner`, `billing_owner`, `designer`, or `implementer`,
- runtime and model configuration for AI actors,
- sandbox posture,
- allowed or denied file scopes,
- configured tools, skills, MCPs, and instructions.

Capabilities are intentionally coarse in the MVP. The policy engine may match the small set it understands first and ignore forward-compatible fields with warnings.

## Actor Selection

The scheduler chooses an actor after node eligibility is known. Selection considers:

- node `work_type`,
- optional requested actor,
- required capabilities from the node or parent contract,
- policy rules,
- risk class and reversibility,
- expected file scope when known,
- runtime availability and concurrency.

A requested actor is a hint. It never bypasses policy.

## Actor Snapshots

Runs and approvals must record actor ids. Runs also record an actor configuration snapshot. Actor definitions will change as prompts, models, tools, MCPs, skills, and roles evolve; historical run audit must preserve what was true when the run started.

The snapshot is not the source of truth for future routing. It is an immutable runtime record for audit, recovery, and later analysis.

## Relationship To Work Types And SOPs

Professional SDLC shape belongs in planner SOPs and the work graph. For example, an organization can require a feature planner to create `value_research`, `functional_spec`, `ux_design`, `technical_design`, `technical_implementation`, `functional_testing`, `deployment`, and `monitoring` nodes with dependencies between them.

Those phases are not statuses. Status remains lifecycle state (`todo`, `in_progress`, `review`, `rework`, `done`, and so on). `work_type` tells Groundwork which SOPs, actors, validations, and policies apply to a node.

