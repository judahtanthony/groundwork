# ADR 0034: Minimal Git-Landing — Groundwork Owns the Landing Commit

Status: Accepted

## Context

All durable state lives in git. `state-model.md`'s master test — "could Groundwork
rebuild this from the files plus git, or is its loss irreversible?" — makes docs,
code, policies, SOPs, and the ticket exports under `.groundwork/tickets/**/ticket.md`
committed truth, while `state.sqlite` and `runs/`/`approvals/` are ignored and
recomputable.

The M2 landing gate ([ADR 0028](0028-gate-evaluation-engine.md)) opens a
`land_to_main` approval but does not itself commit to git;
[ADR 0027](0027-run-lifecycle-and-checkpoint-records.md) places real git operations
(per-run worktree branches, WIP checkpoints, squash, resume) in Phase 4. So a node
that gw reports as `landed`/`done` is not yet *committed* — the system of record can
drift from the durable truth, and a human is relied on to remember `git commit`.

## Decision

**gw owns the landing commit, via a minimal `internal/git`.** When the
`land_to_main` gate completes (auto-approval, override, or an approved human gate),
the coordinator regenerates the ticket export (now `status: done`), stages it, and
commits on the **current branch** with a ticket-referencing message, returning the
commit SHA (recorded on the audit trail as a `ticket.committed` event). The commit
bundles the human's change and the updated export.

**The store transition and the commit are not one transaction** — SQLite and git
cannot be committed atomically. The node is transitioned to `done` first, then
committed. If the commit fails, the node is recorded `done` but uncommitted, and the
error says so and names the recovery: re-running `gw ticket land <id>` finishes the
commit. The coordinator treats an already-`done` node as a commit-completion request
(idempotent: it re-stages the export and commits only if something is pending), so a
transient git failure is recoverable in-product rather than stranding the node.

**The git index is the ticket-scoped pathspec.** The human stages the files they
changed for the ticket (ordinary `git add`); gw stages the regenerated export and
commits the index. The staging area is thus the human's explicit declaration of what
belongs to this landing, so unrelated *unstaged* edits in the working tree are never
captured. If nothing is staged and the export is already current, gw records the
landing without forcing an empty commit.

Because day-to-day development usually commits everything at once (`git commit -a`),
`gw ticket land` also offers that ergonomic, resolved CLI-side before the coordinator
commits: `--all` stages every change (`git add -A`, honoring `.gitignore`), and when
nothing is staged but the work tree has changes the command asks whether to include
them all. The decision only stages — the coordinator still performs the single commit
— so the staging persists across the land→approve gate.

**The empty-index default depends on isolation.** A present human gets default-yes
(Enter accepts staging everything). But a *non-interactive* caller — EOF / piped
stdin / CI — gets **no**: in M3 landings run against the **shared working tree**, so
auto-staging everything unattended could sweep in unrelated work. This default is a
function of isolation, not a fixed value: when Phase 4 runs a ticket in its own
worktree, the sandbox holds only that ticket's work, so the non-interactive default
flips to add-all — which is precisely the seam progressive autonomy plugs into. The
M3 shared-tree path takes the safe default; the Phase 4 worktree runtime owns the
add-all-by-default behavior, justified by isolation rather than hardcoded here.

The capability is deliberately minimal — **stage + commit on the current branch
only**. It excludes isolated worktrees, branch creation, WIP checkpoints, squash, and
resume, all of which remain Phase 4
([ADR 0027](0027-run-lifecycle-and-checkpoint-records.md)) because they serve
autonomous agent execution, not landing. It refuses to commit on a **detached HEAD**
(the commit would be an orphan, lost when HEAD moves). When the project root is not a
git work tree, gw degrades gracefully: the export is still regenerated and the landing
recorded in the store, only the commit is skipped, so Groundwork runs in non-git
directories and tests without the durable export drifting from store status.

## Consequences

Landing produces a git commit as part of the gate completing, so in the normal case
the self-hosting story needs no manual commit step and gw's recorded state stays
consistent with git. It is *best-effort consistent*, not transactional: a commit
failure leaves the node `done` but uncommitted, surfaced with a one-command recovery
(re-running `gw ticket land`). This introduces `internal/git` earlier than the
`architecture-map.md` phasing (which pencils it into Phase 4) — a documented deviation
scoped strictly to landing. Phase 4's worktree runtime builds the richer git layer
(checkpoint commits on a run branch, squash-to-landing) *on top* of this; the
landing-commit semantics defined here are the target those checkpoints squash into.

For human-performed M3 work the human edits in the shared working tree and the
ticket-scoped pathspec is the git index — what the human chose to stage (or `--all`).
There is deliberately **no guard against unrelated *staged* files**: the staging area
is the human's explicit declaration of intent, so honoring it (rather than second-
guessing it) is the contract. The only protections are structural: unrelated
*unstaged* edits are never captured, the non-interactive empty-index default is `no`
in the shared tree, and commits are refused on a detached HEAD.
