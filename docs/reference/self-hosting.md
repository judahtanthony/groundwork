# Self-hosting runbook (M3)

Groundwork manages its own work. The bootstrap work tree
(`docs/plan/work-tree.yaml`) is imported as managed tickets, and low-risk
documentation changes are taken through Groundwork itself — as the system of record
and the gate keeper — with human landing approval. Agent execution is
human-performed in M3; the Codex runtime is Phase 4.

This runbook is the concrete procedure. For the work-type conventions see
`.groundwork/sops/documentation/SOP.md`; for the decisions behind it see ADRs
0032–0035.

## One-time setup

```sh
gw init                      # scaffold .groundwork (config, actors, policies, WORKFLOW)
go run scripts/bootstrap_import.go   # transcribe work-tree.yaml -> ticket exports
gw ticket import             # rebuild the managed tree from the exports
gw ticket tree               # confirm the hierarchy and statuses
```

`gw` never reads the planning YAML; the durable artifacts are the committed
Markdown exports under `.groundwork/tickets/` (ADR 0032). `state.sqlite` is
runtime-only and git-ignored — it rebuilds from the exports on cold start.

## Run a documentation ticket (human-performed)

Start the coordinator without the scheduler, so the human owns the node lifecycle
rather than the scheduler auto-claiming eligible nodes for an AI actor (ADR 0033):

```sh
gw server --no-scheduler
```

Then, for a `work_type: documentation` node `<id>`:

```sh
gw ticket context <id>             # read the brief (spine, contract, acceptance, SOP)
gw ticket transition <id> in_progress
#   ... edit the docs in the working tree ...
git add <changed-docs>             # stage the ticket-scoped pathspec
gw ticket transition <id> review
gw ticket land <id>                # opens the land_to_main approval (human-gated)
gw approval approve <approval-id>  # gw commits the change + export, node -> done
```

When the landing approval is granted, Groundwork enforces the `documentation`
validation template and then makes the durable commit itself: the human's staged
change plus the regenerated ticket export (now `status: done`) land in one commit on
the current branch (ADR 0034). "Landed" therefore means "committed."

## Verify

```sh
git log -1 --stat        # one landing commit: the doc + .groundwork/tickets/<id>/ticket.md
gw ticket show <id>      # status: done
git status               # no runs/, approvals/, or state.sqlite* staged
```

Runtime state stays ignored; only durable canon is committed.

## Feedback loop

If, while working a node, you needed something the brief did not provide, record a
context-miss so the brief and SOPs improve (ADR 0035, ADR 0013):

```sh
gw ticket context <id> --miss "what you needed but the brief lacked"
```

Recurring misses are promoted into an SOP, a doc, or the context-brief assembly —
that is the canon-as-memory loop the dogfooding validates.
