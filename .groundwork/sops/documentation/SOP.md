# SOP: Documentation work

Operating procedure for `work_type: documentation` nodes in Groundwork. Docs work
is low-risk and reversible, so it is the first work type Groundwork self-hosts
(M3). Execution is human-performed in M3 — Groundwork tracks, gates, validates, and
lands; it does not write the docs (the Codex runtime is Phase 4).

## Scope

Use this SOP for changes whose deliverable is documentation: Markdown under `docs/`,
`AGENTS.md`, `README.md`, `.groundwork/**/*.md` (WORKFLOW, SOPs), and ADRs. Code
changes are out of scope even when they touch comments.

## Procedure (human-performed)

1. **Claim the node.** `gw ticket transition <id> in_progress`. Read the brief first:
   `gw ticket context <id>` gives the ancestor spine, parent contract, acceptance,
   and this SOP.
2. **Do the work in the working tree.** Edit the target docs directly. Keep the change
   to one verifiable unit that satisfies the node's acceptance criteria. Match the
   surrounding document's voice and structure; do not restructure unrelated sections.
3. **Stage exactly the files this node changes** — `git add <paths>`. The staging area
   is the ticket-scoped pathspec: at landing Groundwork commits the index plus the
   regenerated ticket export, so unstaged edits are never captured (ADR 0034).
4. **Hand off for review.** `gw ticket transition <id> review`.
5. **Land through the gate.** `gw ticket land <id>` opens the `land_to_main` approval.
   Landing to `main` stays human-required in v1, so approve it explicitly:
   `gw approval approve <approval-id>`. Groundwork enforces the `documentation`
   validation template, then commits the change + export on the current branch and
   moves the node to `done`.
6. **Verify.** `git log -1` shows one landing commit (doc + export); `git status` shows
   no `runs/`, `approvals/`, or `state.sqlite*` staged — runtime state stays ignored.

## Validation

The `documentation` validation template (`.groundwork/policies/validation.yaml`)
matches `**/*.md`, `AGENTS.md`, `MEMORY.md` with `landing_risk_floor: low` and no
required commands. Prose has no automated check in v1; correctness is the reviewer's
judgement against the acceptance criteria.

## Canon and the feedback loop

Documentation *is* canon (ADR 0013). When a change names a canonical home — an ADR,
an architecture or contract doc, a policy, or this SOP — edit that document in place
rather than leaving the knowledge in a node journal. If, while working a node, you
needed something the brief did not give you, record a context-miss so the brief and
this SOP improve over time (ADR 0035).
