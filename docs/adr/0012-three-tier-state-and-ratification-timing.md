# ADR 0012: Three-Tier State And Ratification Timing

Status: Accepted
Implemented: Partial

## Context

ADR 0002 made SQLite the operational store and ADR 0007 said runtime state is not committed. The shorthand "SQLite = runtime, files = durable" is the consequence, not the rule, and it left a real question unanswered: where does *distilled design knowledge* produced during a run belong? It is produced by machinery (a run) but, once settled, it is project intent. Treating "where it was produced" as the deciding factor leads to either stranding design rationale in the disposable store or scribbling every in-progress thought into git.

## Decision

The deciding question for any datum is: **could Groundwork rebuild this from the files plus git, or is its loss irreversible?** SQLite may hold only what is recomputable or safely lost. Anything whose loss is irreversible must be a file. State falls into three tiers:

- **Ephemeral runtime** (SQLite + ignored run logs): leases, PIDs, live waits, transcripts, raw output, generated views, and in-progress/candidate decisions. Recomputable or safely lost.
- **Durable operational records** (SQLite-primary, exported to files): nodes, statuses, timelines, approval decisions, run manifests. Mutated live, so they need transactions/queries; exported so they survive and rehydrate SQLite on cold start.
- **Canonical knowledge** (file-authoritative, committed): code, docs, ADRs, policy, SOPs, and distilled design. The file is the source of truth; SQLite at most indexes it.

A datum's tier is set by the role it plays once it exists, not by where it was produced. Distilled design is canonical (tier 3); the run that produced it is ephemeral (tier 1).

**Timing — the ratification gate.** Distilled design is written to a file at the moment a decision becomes binding on other work: a decomposition proposal is accepted, a node lands, or a policy/SOP change is approved. Before that boundary it is freely mutable tier-1 state with no repo cost; a decision that is never ratified never touches the repo. Promotion is neither continuous nor deferred to root completion — it happens at each ratification gate.

SQLite and files cooperate rather than compete: SQLite is the index/graph (which contract/ADR/SOP applies to which node, the tree, the dependency edges); files hold the content. This is what lets context be queried per-node while design stays durable (see ADR 0013).

## Consequences

`state-model.md` and `overview.md` are restructured around the three tiers and the master test. The distillation/promotion mechanism and its parent-reconciliation step are specified in ADR 0013. This rule resolves every future "which store?" question by pointing at reconstructability and the ratification gate.
