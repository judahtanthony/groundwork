# Self-hosting runbook (M3)

Groundwork manages its own work: the work tree is the planning source of truth
(ADR 0040). Low-risk documentation changes are taken through Groundwork itself — as
the system of record and the gate keeper — with human landing approval. Agent
execution is human-performed in M3; the Codex runtime is Phase 4.

This runbook is the concrete procedure. For the work-type conventions see
`.groundwork/sops/documentation/SOP.md`; for the decisions behind it see ADRs
0032–0035 and `.groundwork/WORKFLOW.md`.

## Setup

The tree is already managed; the durable artifacts are the committed Markdown
exports under `.groundwork/tickets/`. `state.sqlite` is runtime-only and git-ignored,
so on a fresh checkout it rebuilds from those exports:

```sh
gw ticket import   # rebuild the store from the committed exports (cold start)
gw ticket tree     # confirm the hierarchy and statuses
```

The original bootstrap transcribed the now-retired `work-tree.yaml` into these
exports; that one-time step lives in git history (ADR 0032, 0040).

## Run a documentation ticket (human-performed)

Start the coordinator:

```sh
gw server
```

The scheduler runs, but it dispatches a node to an AI actor only when the trust
policy's `allow_claim` authorizes one. In M3 the project authorizes no AI claims
(`allow_claim: []` in `.groundwork/policies/trust.yaml`), so the scheduler leaves
every node for the human and never races for it (ADR 0033). Handing work to the
scheduler later is a policy change — add an `allow_claim` rule for the work type —
not a server mode.

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

Stage just the ticket's files (above) to keep the commit ticket-scoped, or use
`gw ticket land <id> --all` to stage every change first (the `git commit -a`
ergonomic). If nothing is staged, the command asks whether to include all changes
(default yes).

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
