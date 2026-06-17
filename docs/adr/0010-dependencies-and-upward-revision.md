# ADR 0010: Dependencies And Upward Revision

Status: Accepted

## Context

Dynamic decomposition (ADR 0009) lets a composite node spawn children. Those children may be safely parallel, or they may depend on one another, and a child may discover that the parent's design is wrong after sibling work has already started.

## Decision

Add **dependency edges** as a directed-acyclic overlay on the work tree. Scheduler eligibility becomes `todo AND all dependencies satisfied`; cycles are rejected. Prefer a complete parent **contract** (shared schemas/interfaces recorded on the parent) so children need no inter-child edges and can run in parallel; fall back to explicit dependency edges to serialize work only when the contract cannot be made complete enough.

Add an **escalation / upward-revision** protocol: a child that finds a design problem records a typed escalation event, transitions to `blocked`, and routes a re-plan decision to its parent. In v1 the parent re-plan and any resulting sibling `rework` are **human-gated**; automatic cascading invalidation across siblings is deferred to Phase 2.

## Consequences

The scheduler and `blocked` semantics extend to dependency-aware eligibility and escalation. Dependency edges and escalation events become durable, exported state. The DAG must be validated for cycles when edges are added.
