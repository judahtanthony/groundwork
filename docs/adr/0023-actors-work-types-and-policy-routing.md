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

## Provisional: Actor Identity Scheme

> **Resolved in Phase 2 by [ADR 0029](0029-actor-identity-model.md):** the identity
> path carries authority/grouping only (prefix-matchable, variable depth),
> capabilities are a separate parameterized property object, and routing matches a
> class while audit records the resolved instance. The provisional notes below are
> retained for historical context.

The actor **identity** model here is intentionally minimal and **provisional**. The MVP uses a closed `type` set and flat actor ids — the degenerate "one instance per capability class" case. Two generalizations are anticipated and should be settled early in Phase 2, when actor-aware matching first binds code to the model (the ratification gate, per [ADR 0013](0013-canon-as-memory.md)):

- **Class vs instance.** An actor *class* is a fungible capability set (e.g. a front-end engineer); an *instance* is the unique entity that holds and performs work — the node's `assignee` — possibly bound to a long-lived environment. Routing matches the class; audit records the instance. A pool of interchangeable instances per class is the scaling shape (humans are already instances-of-a-role; mature agents follow the same form).
- **Tiered identity.** Identity may become a hierarchical, dotted namespace where any prefix is matchable: `human` → `human.frontend` → `human.frontend.judahtanthony`, or `agent` → `agent.backend-go` → `agent.backend-go.instance-a3fe35`. A policy can start coarse ("a `human` must approve") and tighten to a specific tier ("only `human.director.judahtanthony` may create new credit cards"). The tier depth a rule matches at is the fungibility boundary. Identity tiers (who / authority / grouping) stay **separate from capabilities** (what tools / work types an actor has): capabilities are a set, not a path.

Until Phase 2 ratifies this, treat `type` as the coarsest tier, actor `id` as a free-form (already dot-segmented) string, and `requested_actor` / `assignee` as forward-compatible placeholders for the class request and the resolved instance. Nothing in v1 is hard-bound to the flat scheme except the `type`-enum validation, which can be relaxed when the tiered model is adopted.

