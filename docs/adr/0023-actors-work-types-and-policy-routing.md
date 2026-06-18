# ADR 0023: Actors, Work Types, And Policy Routing

Status: Accepted

## Context

Groundwork needs to support a solo developer running a largely AI-driven local software factory while leaving a path to richer human-agent collaboration. Professional software delivery also has organization-specific workflow shape: discovery, functional definition, UX, technical design, implementation, review, testing, deployment, and monitoring vary by team.

Encoding those phases as ticket statuses would make the status model organization-specific and brittle. Encoding them as hardcoded node kinds would also limit planner SOPs. Groundwork already has a uniform work-node tree and dependency DAG, which is the better place to model process-specific work.

The system also needs to distinguish actors. A human owner, a Codex implementer, an AI documentation judge, and a future billing-domain human reviewer should be policy-distinct even if they all act on the same work graph.

## Decision

Groundwork keeps status lifecycle-oriented and adds operational routing metadata:

- `work_type` on work nodes, defined by planner SOPs and organizations.
- `.groundwork/actors.yaml`, a committed local registry of human and AI actors.
- Actor-aware policy matching for claims, approvals, reviews, edits, landing, and gated actions.
- Actor ids and actor configuration snapshots on runs.

The MVP registry contains `human.owner` and `ai.codex.default`. Actors may declare type, roles, capabilities, runtime, model, sandbox, and coarse limits. Policy may match actor id, actor type, roles, work type, action, files, risk class, and reversibility.

SDLC shape belongs in planner SOPs and generated graph structure. A planner may create nodes such as `value_research`, `functional_spec`, `ux_design`, `technical_design`, `technical_implementation`, `functional_testing`, `deployment`, and `monitoring` with dependency edges between them. Those are work types and nodes, not statuses.

## Consequences

The MVP remains small: one local owner, one default Codex actor, local policies, and no account system. This is consistent with the single-human-operator scope of [ADR 0005](0005-localhost-single-user-v1.md): multiple *actors* (including AI actors) act on the graph in v1, while multiple *human* roles with authentication arrive only with the post-v1 remote/LAN mode. At the same time, the data model can later support multiple human roles, task-specific AI actors, AI judges, several runtimes, and multiple LLM providers without changing the core status model.

Policies become the authorization and routing layer. Requested actors on tickets are hints, not authority. Runs snapshot actor configuration so audit remains stable after actor definitions change.

The status enum should not grow to include organization-specific SDLC phases. `rework` remains a lifecycle state for failed review with actionable feedback; newly discovered scope should become new dependent nodes or an upward re-plan.

