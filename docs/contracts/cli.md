# CLI Contract

The CLI name is `gw`.

## Required Commands

```text
gw init
gw init --update-agents-md
gw ticket create
gw ticket list [--status <s> | --ready | --blocked]
gw ticket show <id>
gw ticket edit <id> [--parent <id>]
gw ticket claim <id> [--actor <id>]
gw ticket assign <id> <assignee>
gw actor list
gw actor show <id>
gw actor validate
gw ticket transition <id> <status>
gw ticket tree [id]
gw ticket context <id>
gw ticket decisions <id>
gw ticket request <id>
gw ticket decompose <id>
gw ticket link <id> --depends-on <id>
gw ticket escalate <id>
gw ticket land <id> [--all] [--override] [--preview]
gw ticket export [id]
gw ticket import [path]
gw next [--claim] [--actor <id>]
gw status
gw board
gw run once <ticket-id>
gw run next
gw run list
gw run show <run-id>
gw run pause <run-id>
gw run resume <run-id>
gw run cancel <run-id>
gw approval list
gw approval show <approval-id>
gw approval approve <approval-id>
gw approval reject <approval-id>
gw approval clarify <approval-id>
gw validation list <ticket-id>
gw validation run <ticket-id>
gw server
gw doctor
gw export
gw sync
```

## Context Brief

`gw ticket context <id>` returns the bounded, node-specific brief an agent receives at claim time: ancestor spine, parent contract, direct dependencies, relevant SOPs, actor constraints, open escalations, relevant durable decision/input records, latest rework notes, validation state, checkpoint/diff refs, and prior handoff summaries. It reads canon and durable ticket records resolved through the SQLite graph; broader queries (for example `--siblings`) are explicit. See [ADR 0013](../adr/0013-canon-as-memory.md) and [ADR 0051](../adr/0051-async-agent-handoff-and-durable-decision-records.md).

`gw ticket decisions <id>` lists the durable ticket-attached decision/input/approval
history. `gw ticket request <id>` is the future command surface for creating a local
input request or durable decision request without pretending that every clarification is
a full work node.

## Eligibility And The Next Node

`gw next` is the **human** picker: it names the top node of the eligible set (todo +
dependencies satisfied, value-ordered per [ADR 0039](../adr/0039-value-prioritization-v1.md)),
prints its brief, and with `--claim` takes it. `gw ticket list --ready` shows the whole
eligible set in the same order; `gw ticket list --blocked` shows todo nodes held back by
unsatisfied dependencies. These read the same surface the scheduler dispatches from
([ADR 0041](../adr/0041-human-cli-operating-model.md)). Distinct from `gw run next`, which
is **machine** dispatch to an AI actor through the policy gate.

## Actors

`gw actor list`, `gw actor show <id>`, and `gw actor validate` operate on `.groundwork/actors.yaml`. Actor definitions are committed project configuration; run history stores actor snapshots separately.

## Output

Default output should be human-readable and script-friendly. Every data command should support `--json`.

## Coordinator Interaction

If `gw server` is running, mutating commands should call the local API by default. If it is not running, simple ticket/config commands may open SQLite directly through the shared store package. Commands requiring live run control must fail clearly without the coordinator.

The scheduler only dispatches a node to an AI actor when the trust policy's `allow_claim` authorizes that actor to claim it. A project that authorizes no AI claims keeps human-performed work's lifecycle free of scheduler interference; loosening `allow_claim` is what makes work available to the scheduler ([ADR 0033](../adr/0033-human-execution-via-manual-transitions.md)).

Approvals remain the CLI for live gated actions. For durable gates, the approval id is a
runtime handle over a stable ticket decision/request id; after a cold rebuild, the
approval id may change while `gw ticket decisions <id>` still shows the semantic record.
