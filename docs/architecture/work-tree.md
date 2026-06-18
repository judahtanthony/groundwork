# Work Tree

Groundwork models work as a hierarchical breakdown of intent (the *what*) into implementation (the *how*). Every item is a uniform **work node**; the tree captures parentage, and a dependency overlay captures ordering.

## Work Nodes

All work is one node type. `kind` is an advisory human label (for example `goal`, `initiative`, `epic`, `feature`, `ticket`, `task`, `checklist_item`) used for navigation and reporting. It carries no structural meaning.

The only structural distinction is **leaf** versus **composite**, and it is decided at claim time by the triage gate rather than fixed when the node is authored.

`work_type` is separate from `kind`. It is operational metadata used by planner SOPs, policy, actor routing, validation, and context assembly. Organizations may define their own work types (for example `value_research`, `functional_spec`, `ux_design`, `technical_design`, `technical_implementation`, `functional_testing`, `deployment`, or `monitoring`) to reflect their SDLC. Groundwork must not encode those SDLC phases as statuses; the graph models the work, while status models lifecycle state.

Nodes may optionally carry a requested actor or required actor capabilities. These are routing hints and policy inputs, not ownership facts. The scheduler still has to match the node, its risk, touched files, requested action, and available actors against policy before a claim or gated action proceeds.

## Triage Gate

When a node is claimed, the agent triages it ("definition of ready"):

- A node is a **leaf** when it needs no further research or design **and** is implementable and verifiable as one unit. A leaf is dispatched to an executing agent (for example a coding agent).
- A node is **composite** when it has ambiguity or can be broken into smaller parts. The agent records the necessary research, design, and requirements, then decomposes the node into child nodes.

Decomposition is just-in-time and agent-driven, but its output is a **proposal**: a composite node enters `review`, and its children are created non-dispatchable (in `backlog`) until the plan is accepted, at which point they move to `todo` as dependencies allow. See [trust-and-approvals.md](trust-and-approvals.md) and [ADR 0009](../adr/0009-generalized-work-nodes-and-dynamic-decomposition.md).

## Parent Contract And Parallelism

The preference is that a parent records a **contract** — schemas, interfaces, and requirements — complete enough that its children share no undefined surface. The completeness of that contract decides how children run:

- **Complete contract:** children have no undefined shared surface and run in parallel with no inter-child dependencies.
- **Incomplete contract:** either push more design into the parent first, or add explicit **dependency edges** so dependent children are serialized.

A leaf still represents one verifiable change with clear acceptance criteria and enough scope for an agent to complete independently.

## Dependencies

Dependency edges form a directed-acyclic overlay on the tree (cycles are rejected). A node is eligible for dispatch only when it is `todo` and all of its dependencies are satisfied, so picking up a node means its prerequisites are already met. See [ADR 0010](../adr/0010-dependencies-and-upward-revision.md).

## Context And Navigation

- An agent claiming a node receives its full **ancestor context** automatically: the chain of parents and the relevant parent contract. It may query siblings or any other node on demand.
- Humans see the parentage breadcrumb on any node, but the primary units navigated are the leaf nodes; parent presentation is largely computed from children.

## Rollups

Parent state is derived from children where possible:

- all children done -> parent done,
- any child blocked -> parent has blocked work,
- active child -> parent in progress,
- no active children and unfinished children -> parent planned or todo.

Manual roadmap states may exist for parent nodes, but executable state belongs to leaves.

## Canon Reconciliation

When children promote durable design into canon, the composite parent **owns reconciling it**. The parent reviews and refines its children's promoted contributions so the result is coherent and non-redundant — no conflicting or repetitive design description. Reconciliation happens at the parent, at the ratification gate, serialized through the coordinator rather than concurrently in worktrees, which also keeps canon writes conflict-free. The parent contract is the forward (promote-on-dependency) channel; this reconciliation plus the completion roll-up are the retrospective channel. See [ADR 0013](../adr/0013-canon-as-memory.md).

## Upward Revision

Revision can propagate up the tree. If a node discovers a mistake or complication that changes the requirements, it records a typed **escalation**, transitions to `blocked`, and routes a re-plan decision to its parent. The parent may adjust its design and plan and may send affected siblings to `rework`. In v1 the re-plan and any sibling rework are human-gated; automatic cascade is deferred. See [ADR 0010](../adr/0010-dependencies-and-upward-revision.md).

## Managing Large Work

Large product work starts as a high-level node, is decomposed (up front or just-in-time) into smaller nodes, and bottoms out in leaves. Runtime history attaches to leaves and runs. Product design and durable decisions attach to parent contracts, specs, and ADRs referenced by parent nodes.
