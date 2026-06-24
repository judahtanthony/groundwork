# ADR 0043: Scalable Human-Agent Collaboration

Status: Draft
Implemented: Partial

## Implementation State

Groundwork already has the seed mechanics this direction builds on: uniform work nodes, dependency-aware scheduling, policy gates, approvals, validation records, canon, and the local actor model. The broader collaboration model described here remains draft future direction and does not override accepted ADRs.

## Goal

Groundwork should make human and agent collaboration as streamlined as possible while
retaining the structure, visibility, and quality controls needed for real software work.
The MVP assumes one engineer and one agent on one machine, but the architecture should
not bake in a single universal human or a single active feature.

The ideal system lets humans approve goals, constraints, and exceptions while agents
progressively decompose, execute, integrate, validate, and summarize work inside visible
boundaries.

See [ADR 0050](0050-agentic-software-factory-direction.md) for the industry
direction behind this collaboration model: background coding agents, agent harnesses,
software-factory workflows, deterministic validation, isolated execution, and governance
around progressively autonomous work.

## Problem

The current design intentionally decomposes complex work into narrow, verifiable nodes.
That improves agent execution quality, but it creates two scaling pressures:

- More small nodes create more approval points for planning, landing, tactical actions,
  escalation, and re-planning.
- Dependency edges prevent logical ordering problems, but they do not fully prevent
  agents from touching the same files, symbols, APIs, fixtures, or product surfaces in
  parallel.

The system also needs to grow from the MVP into multiple human roles. A product owner
may approve root goals and PRDs, while a staff engineer or CI owner may approve changes
to build, release, or infrastructure files. The MVP can let one engineer satisfy every
role, but policy should already understand which role was required.

## Design Direction

Groundwork should move human review upward from every internal step to the boundaries
where judgment matters:

- root or parent goals,
- approved planning envelopes,
- scope expansions,
- cross-root conflicts,
- failed or missing validation,
- high-risk files or actions,
- final feature-level outcomes,
- policy and authority changes.

Agents should handle bounded execution. Groundwork should prove that each run stayed
inside its approved contract, passed required validation, and produced a concise summary
before the system reduces review load.

## Operating Model

1. A human creates or approves a root node that represents a feature-level or outcome-level
   change.
2. The planner proposes an initial decomposition, dependency graph, resource scope, risk
   classification, validation template, and candidate actors.
3. The reviewer approves a planning envelope, not necessarily every future child.
4. The scheduler runs eligible, non-conflicting children up to the worker-pool limit.
5. Each child executes in an isolated worktree and records events, diffs, validation,
   and a completion summary.
6. Parent nodes integrate child outputs, resolve conflicts, update parent memory, and run
   broader validation.
7. Low-risk in-scope child landings can be batched, reviewer-agent checked, or eventually
   auto-approved by policy.
8. Human review focuses on final root outcomes and exceptions.
9. Durable lessons are distilled into canon: docs, ADRs, contracts, policies, and SOPs.

## Approval Principles

- Approvals unlock actions, not people.
- The required approver should be expressed as a role, capability, or actor policy
  requirement.
- The MVP local owner may satisfy every role, but approval records should still state
  which role was required and why.
- Human approval requirements are conservative defaults. Auditability, default-deny,
  validation, reversibility, and policy explanation are the invariants.
- Review load should be reduced through approved envelopes, validation, scope checking,
  reviewer agents, batching, and sampling before full autonomy.

## Parallelism Principles

Groundwork should use two orthogonal controls:

- The dependency DAG answers "what must happen first?"
- Resource scope and ownership answer "what may run at the same time?"

Multiple root nodes may be active at once. Each root should have an integration target,
and the scheduler should consider global resource reservations before dispatching children
from different roots in parallel.

## Related Draft ADRs

- Hierarchical planning and approval envelopes.
- Node branching and parent integration.
- Resource scope, ownership, and conflict policy.
- Hierarchical memory and context compression.
- Role-aware actors and local identity.
- Full-SDLC work modeling.

## Proposed Work Breakdown

- Add draft contracts for planning envelopes and resource scopes.
- Extend policy matching to include planned scope, actual scope, required roles, and
  approval target.
- Add local identity resolution for "my work" commands and web views.
- Add parent integration branch semantics before enabling broad autonomous child landing.
- Add completion summaries and parent memory assembly.
- Add web-admin queues grouped by required role, parent, risk, and exception reason.
- Add harness/blueprint execution so SOPs can become partially executable workflows.
- Add background run triggers after the local coordinator can safely isolate, validate,
  observe, and govern agent runs.
