# ADR 0052: Consequential Decisions As Work Nodes

Status: Accepted
Implemented: Not started

## Context

Autonomous agents will encounter questions that are not just local clarification.
Some decisions change architecture, scope, policy, parent contracts, dependency
ordering, or cross-ticket coordination. Treating every question as a chat prompt
keeps the original worker alive, hides ownership and routing, and makes the
result hard to validate or promote to canon. Treating every uncertainty as a
ticket would be noisy and slow.

Groundwork already has the substrate needed for meaningful decisions: uniform
work nodes, dependencies, `work_type`, actor routing, policy gates, validation,
and canon promotion.

## Decision

Consequential decisions are work. Use normal Groundwork work nodes and dependency
edges for decisions whose answer has independent scope, ownership, validation, or
canon impact.

Use decision tickets for:

- architectural decisions,
- scope or parent-contract changes,
- risky or out-of-affordance choices,
- cross-ticket coordination,
- decisions requiring research,
- decisions needing a specialist actor or model,
- decisions that should produce canon,
- decisions that block multiple nodes.

Do not create tickets for every small uncertainty. Use local input requests for
bounded clarifications needed only to continue the current run, with no
independent validation, canon impact, or parent-contract change. Use approval
gates for permission on a concrete action such as landing, executing a risky
command, or accepting/rejecting a concrete proposal.

The ladder is:

```text
run event -> input request -> approval gate -> decision ticket -> canon
```

Decision nodes remain normal work nodes:

- `kind` may be `decision` for navigation, but it is advisory.
- `node_type` is still `leaf` or `composite`.
- `work_type` drives routing, policy, SOP, and validation.
- dependency edges connect blocked implementation work to the decision.

Suggested work types include:

- `architecture_decision`
- `scope_clarification`
- `risk_review`
- `policy_exception`
- `dependency_resolution`
- `technical_design`

Example:

```text
T-1204 Implement resume packet
  depends_on:
    - T-1211 Decide async memory authority model

T-1211
  kind: decision
  node_type: leaf
  work_type: architecture_decision
  requested_actor: ai.architect.high_context
  acceptance:
    - ADR records the decision.
    - Impacted architecture/contracts are updated.
    - Blocked implementation ticket can proceed.
```

If an agent exits because work is blocked on a durable decision, the decision
node, dependency edge, and ticket-attached decision/request record must be
exported before the run is considered safely blocked.

## Consequences

No separate decision subsystem is introduced. The scheduler, policy engine,
context assembler, validation templates, and canon distillation all operate over
the same work tree. This preserves ADR 0036's work-as-substrate model while
keeping the threshold explicit enough to avoid turning every question into a
ticket.

`work-tree.md`, `runtime-model.md`, `cli.md`, `http-api.md`, and
`conventions.md` are updated to describe the threshold and routing behavior.
