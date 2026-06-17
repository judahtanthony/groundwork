# Architecture Overview

Groundwork is a local coordination system around one repository.

Conceptual flow:

```text
gw CLI / Web UI / Agent Runner
  -> Coordinator
  -> SQLite operational store
  -> Exporters
  -> .groundwork files
```

The coordinator owns live decisions. SQLite gives transactional claims, leases, approvals, run state, validation gates, dependency eligibility, decomposition proposals, and escalations. Exported files make durable project state inspectable by humans and agents.

## Operational State Versus Durable State

State falls into three tiers, decided by one test: could Groundwork rebuild this from files plus git, or is its loss irreversible? See [ADR 0012](../adr/0012-three-tier-state-and-ratification-timing.md) and [state-model.md](state-model.md).

- **Ephemeral runtime** (SQLite + ignored run logs): leases, PIDs, live waits, transcripts, command logs, per-node journals, approval-inbox projections, generated views, worktree contents before landing. Recomputable or safely lost.
- **Durable operational records** (SQLite-primary, exported to files): nodes, statuses, timelines, approval decisions, run manifests.
- **Canonical knowledge** (file-authoritative, committed): code, docs, ADRs, trust/risk/validation/autonomy policies, SOPs, and distilled design promoted at the ratification gate ([ADR 0013](../adr/0013-canon-as-memory.md)).

SQLite is the runtime source of truth and the index/graph; files hold durable content and are not committed for ephemeral tiers. Exporters project durable operational records to files; canonical knowledge is authored directly as files.

## Boundaries

The first implementation should keep these subsystem boundaries:

- CLI and HTTP handlers call coordinator/store services.
- The scheduler owns claims and run lifecycle.
- Agent runtimes communicate through a runtime interface.
- Policy and validation engines produce decisions, not side effects.
- Exporters write readable projections from canonical state.

