# ADR 0009: Generalized Work Nodes And Dynamic Decomposition

Status: Accepted

## Context

The original work tree used a fixed taxonomy (`goal → initiative → epic → feature → ticket → task → checklist_item`) in which `ticket` was a specific executable kind, and decomposition was assumed to be a human activity performed up front. In practice the right breakdown is often only discoverable when work is picked up, and the unit an agent should execute is not tied to a taxonomy label.

## Decision

Model all work as a single uniform **work node**. `kind` becomes an advisory human label, not a structural property. The only structural distinction is **leaf** versus **composite**, decided by a **triage gate** ("definition of ready") when the node is claimed:

- A node is a **leaf** when it needs no further research or design and is implementable and verifiable as one unit. It is dispatched to an executing agent.
- Otherwise it is **composite**. The agent records the research, design, and requirements it discovers and decomposes the node into child nodes just-in-time.

Decomposition output (research, design/contract, and child nodes) is a **proposal**. A composite node enters `review` and its children are not dispatchable until the plan is accepted. Parent state continues to roll up from children.

## Consequences

This collapses the fixed taxonomy in `work-tree.md` to advisory metadata plus the leaf/composite structural fact. Agents gain a plan-mutation capability that must be gated (see ADR 0011). Runs split into planning/decomposition runs and implementation runs (see `runtime-model.md`).
