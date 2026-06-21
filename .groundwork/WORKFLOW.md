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
cold start.

## The loop

1. **Orient.** Read AGENTS.md (boundary + invariants/defaults), then `gw ticket tree`
   for the live plan and `gw ticket list --status todo` for what is ready.
2. **Pick the next node.** The eligible set (todo + dependencies satisfied) is ordered
   by value — priority down the ancestor path, then DFS/FIFO (ADR 0039). Take the top.
3. **Read the brief.** `gw ticket context <id>` — ancestor spine, parent contract,
   acceptance, dependencies, and the relevant SOP.
4. **Triage.** A leaf is one verifiable change → execute it. A composite →
   `gw ticket decompose` into a reviewable child proposal (children land in backlog
   until accepted).
5. **Execute.** Human-performed via manual transitions
   (`gw ticket transition <id> in_progress` → … → `review`) in the current phase; an
   AI actor is dispatched only when the trust policy `allow_claim` authorizes one
   (ADR 0033) — handing work to agents is a policy change, not a mode.
6. **Validate + land.** `gw ticket land <id>` opens the `land_to_main` gate (human
   approval in v1); approving it enforces the validation template and the coordinator
   makes the durable git commit — "landed" means "committed" (ADR 0034).
7. **Distill + feed back.** When a change names a canonical home (an ADR, doc, policy,
   or SOP), edit that document in place. Record anything the brief lacked with
   `gw ticket context <id> --miss "…"` so canon and the briefs improve (ADR 0013/0035).
8. **Re-plan.** If a node uncovers a mistake, `gw ticket escalate <id>` routes an
   upward revision; the plan is living state, not a static file.

## Conventions

- Keep leaf nodes to one verifiable change.
- Use `work_type` for SOP / routing / validation / policy semantics; do not encode an
  SDLC into the status model.
- Set `priority` (`[0,1]`) on a node to order it ahead of its siblings; it orders the
  eligible set, never authorizes (ADR 0039).
- Dependency edges form a DAG; a node is eligible only when `todo` and all
  dependencies are satisfied.
- Authority (landing, decomposition, irreversible actions, elevation) is a loosenable
  policy gate, conservative by default; never self-elevate (ADR 0037/0038).
