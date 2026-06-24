# Architecture Overview

Groundwork is a local coordination system around one repository.

Conceptual flow:

```text
gw CLI / Web UI / Agent Runner
  -> Coordinator
  -> .groundwork files
  -> SQLite projection / runtime store
```

The coordinator owns live arbitration. SQLite gives transactional claims, leases, live
approval/input queue projections, run state, validation gates, dependency eligibility,
actor snapshots, and escalations. Exported files make durable project state inspectable
by humans and agents. Ticket sidecar decision records are the durable home for
rebuildable blockers, proposals, inputs, approvals, rework requests, and recovery states
(ADR 0051).

## Operational State Versus Durable State

State falls into three tiers, decided by one test: could Groundwork rebuild this from files plus git, or is its loss irreversible? See [ADR 0012](../adr/0012-three-tier-state-and-ratification-timing.md) and [state-model.md](state-model.md).

- **Ephemeral runtime** (SQLite + ignored run logs): leases, PIDs, live queue handles, transcripts, command logs, per-node journals, approval-inbox projections, generated views, worktree contents before landing. Recomputable or safely lost.
- **Durable operational records** (file-authoritative, projected into SQLite): nodes, statuses, timelines, ticket-attached decision/input/gate records, approval decisions, run manifests.
- **Canonical knowledge** (file-authoritative, committed): code, docs, ADRs, trust/risk/validation/autonomy policies, SOPs, and distilled design promoted at the ratification gate ([ADR 0013](../adr/0013-canon-as-memory.md)).

SQLite is the live transactional index/graph and runtime store; files hold durable
content and are not committed for ephemeral tiers. Mutations to durable state must write
through to files, or to an explicit durable replay record, before reporting durable
success (ADR 0053). Canonical knowledge is authored directly as files.

## Boundaries

The first implementation should keep these subsystem boundaries:

- CLI and HTTP handlers call coordinator/store services.
- The scheduler owns claims and run lifecycle.
- Actor registry and policy determine which humans or agents may claim, review, approve, edit, or land work.
- Agent runtimes communicate through a runtime interface.
- Policy and validation engines produce decisions, not side effects.
- Exporters write readable projections from canonical state.
