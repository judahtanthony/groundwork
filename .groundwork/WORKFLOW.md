# Workflow

The committed agent operating contract for Groundwork in this repository. It
applies to every node; work-type specifics live in `.groundwork/sops/<work-type>/`.

## Source of truth

The **gw work tree is the plan** — what work exists, its ordering, status, and
acceptance (ADR 0040). Inspect and evolve it through `gw`, not by editing static
files. **Canon** — `docs/` (product, architecture, contracts), ADRs, SOPs, and
policies — is the durable *why/what* a completed subtree leaves behind (ADR 0013).
gw holds the work to evolve canon; canon files hold the result.

`.groundwork/state.sqlite` is runtime-only and git-ignored; the durable plan is the
committed Markdown exports under `.groundwork/tickets/`, which rebuild the store on
cold start. Ticket sidecar decision records under each ticket directory are the durable
home for blockers, input/approval requests, rework notes, and recovery states that must
survive rebuild; SQLite queues are live projections. A ticket mutation is not durable
unless the filesystem export, or a durable replay record, has been written (ADR 0053).

## The loop

1. **Orient.** Read AGENTS.md (boundary + invariants/defaults), then `gw ticket tree`
   for the live plan and `gw next` (top recommended node) or `gw ticket list --ready`
   (the whole eligible set) for what is workable now. Prefer these over
   `gw ticket list --status todo`, which ignores dependencies and so lists blocked
   nodes; `gw ticket list --blocked` shows what is waiting and on what.
2. **Pick the next node.** The eligible set (todo + dependencies satisfied) is ordered
   by value — priority down the ancestor path, then DFS/FIFO (ADR 0039). `gw next` names
   the top node and prints its brief; take it.
3. **Read the brief.** `gw ticket context <id>` — ancestor spine, parent contract,
   acceptance, dependencies, relevant SOP, durable decision/input records, blockers,
   validation state, checkpoint/diff refs, and prior handoff summaries.
4. **Triage.** A leaf is one verifiable change → execute it. A composite →
   `gw ticket decompose` into a reviewable child proposal (children land in backlog
   until accepted).
5. **Execute.** Human-performed via transitions: `gw ticket claim <id>` is the guided
   one-step start (verifies eligibility, assigns, moves to `in_progress`, prints the
   brief); `gw next --claim` claims the top node directly. From there continue with
   `gw ticket transition <id> review` (the raw `gw ticket transition` remains available).
   An AI actor is dispatched only when the trust policy `allow_claim` authorizes one
   (ADR 0033) — handing work to agents is a policy change, not a mode.
6. **Validate + land.** Stage the ticket's files, optionally preview the change set with
   `gw ticket land <id> --preview`, then `gw ticket land <id>` opens the `land_to_main`
   gate (human approval in v1); approving it enforces the validation template and the
   coordinator makes the durable git commit — "landed" means "committed" (ADR 0034).
7. **Distill + feed back.** When a change names a canonical home (an ADR, doc, policy,
   or SOP), edit that document in place. Record anything the brief lacked with
   `gw ticket context <id> --miss "…"` so canon and the briefs improve (ADR 0013/0035).
8. **Re-plan.** If a node uncovers a mistake, `gw ticket escalate <id>` routes an
   upward revision; the plan is living state, not a static file.

For autonomous/background work, do not keep an agent alive waiting on a human. If blocked,
checkpoint if applicable, write run evidence, export a durable ticket-attached
request/handoff record, optionally create a dependent decision ticket, move the original
ticket to an explainable blocked state, release the lease, and let the scheduler work on
something else.

## Conventions

- Keep leaf nodes to one verifiable change.
- Use `work_type` for SOP / routing / validation / policy semantics; do not encode an
  SDLC into the status model.
- Set `priority` (`[0,1]`) on a node to order it ahead of its siblings; it orders the
  eligible set, never authorizes (ADR 0039).
- Dependency edges form a DAG; a node is eligible only when `todo` and all
  dependencies are satisfied.
- Meaningful decisions are work when they have scope, policy routing, dependency impact,
  validation, or canon output. Small clarifications stay as local input requests.
- Authority (landing, decomposition, irreversible actions, elevation) is a loosenable
  policy gate, conservative by default; never self-elevate (ADR 0037/0038).
