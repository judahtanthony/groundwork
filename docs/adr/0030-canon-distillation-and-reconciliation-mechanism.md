# ADR 0030: Canon Distillation And Reconciliation Mechanism

Status: Accepted
Implemented: Partial

## Context

[ADR 0013](0013-canon-as-memory.md) decided *that* durable design is distilled
from the per-node journal into canon at the **ratification gate**, with the
composite parent reconciling its children's promoted design, all serialized
through the coordinator. [ADR 0012](0012-three-tier-state-and-ratification-timing.md)
fixed the *timing*. Neither pins the *mechanism*, and the M1 work-tree ticket for
it (T-0306) was deferred out of Phase 1. Phase 2 introduces the approval
decisions that *are* the ratification gates, so the mechanism must be built — but
the agent that authors rich distillation content is the Codex runtime (Phase 4).
This ADR specifies what M2 builds versus defers.

## Decision

Build the **distillation plumbing and the forward (promote-on-dependency)
channel** in M2; the agent-authored retrospective content arrives with the
runtime.

- **Ratification-gate hooks.** Distillation is triggered only at ratification
  boundaries, implemented as a hook on the relevant approval decisions and
  landing: (a) a `decompose` proposal is accepted, (b) a node lands, (c) a
  policy/SOP change is approved. No continuous writing; a decision never ratified
  never touches the repo ([ADR 0012](0012-three-tier-state-and-ratification-timing.md)).
- **Journal location.** The per-node journal is tier-1 ephemeral, one file per
  node so parallel runs never conflict, under the ignored run-log tree
  (`run-logs.md`). M2 defines its path and append API; it is never committed.
- **Promote-on-dependency (forward channel) — built in M2.** Accepting a
  decomposition writes the **parent contract** (schemas/interfaces/requirements)
  into canon via the composite node's exported `## Design / Contract` section and
  creates children in `backlog`, moving them to `todo` as dependencies allow
  (`work-tree.md`, `ticket-export.md`). This is the channel siblings read through
  `gw context`, so it must exist for the decompose flow (T-0430) to be meaningful.
- **Serialized canon writes.** All canon writes go through the coordinator on the
  scheduler's serial path, never concurrently from worktrees, which keeps canon
  conflict-free ([ADR 0013](0013-canon-as-memory.md)). Writes are **typed
  promotion**: edit the named canonical document in place (replace, not append)
  so repo size tracks current design; content with no canonical home stays in the
  journal.
- **Parent reconciliation — mechanism in M2.** When children promote design, the
  composite parent owns reconciling it into a coherent, non-redundant whole at
  ratification (`work-tree.md`). M2 builds the reconciliation step and its
  serialization; the *semantic merge* of free-text design is exercised via the
  stub and becomes substantive when the runtime authors journal content.
- **Distill-on-completion (retrospective channel) — hooks in M2, content in
  Phase 4.** The lossy roll-up that compacts outcomes as a subtree closes is
  wired as a landing hook now, but produces meaningful prose only once the Codex
  runtime writes journals. M2 verifies the plumbing (hook fires, write is
  serialized, nothing is written without ratification), not authored content.

## Consequences

`internal/canon` owns the journal API, the ratification hooks, typed promotion,
and parent reconciliation, invoked by the approval/landing flow. `gw context`
(Phase 1) is the read side of the same loop and is unchanged. The retrospective
content and journal authorship depend on Phase 4; M2 delivers the
forward/parent-contract channel and the serialized write path so decomposition
and landing are correct and conflict-free. Fulfills work-tree T-0306.
