# ADR 0058: Integration Targets And Landing Levels v1

Status: Accepted
Implemented: Partial

Concretizes the git branching model of [ADR 0045](0045-node-branching-and-parent-integration.md)
into a minimal, pre-runtime v1, and generalizes [ADR 0034](0034-minimal-git-landing.md)
(commit-to-current-branch) into a target-aware model. It is the git substrate the Phase 5
envelope/landing ADRs assume: [ADR 0054](0054-approval-envelopes-v1.md) (integration
target), [ADR 0056](0056-envelope-aware-claim-authorization.md) (`land_to_parent`), and
[ADR 0057](0057-bulk-review-bundles-v1.md) (root landing review).

## Context

"Landing" and roll-up only mean something concrete once tied to git. The other Phase 5
ADRs use "integration target" and `land_to_parent` without specifying the branch
mechanics; today only ADR 0034 (commit to the current branch) and ADR 0045 (draft,
worktree model deferred to Phase 6) exist. This ADR fixes the v1 mechanics and separates
the durable contract from the evolving implementation (ADR 0037).

## Durable Contract — Landing Levels

Independent of git mechanics, landing is hierarchical:

- **child → parent**: integrate a child result into its parent's integration target.
- **parent → parent**: integrate a completed composite upward.
- **root → main**: final feature-level review and merge — **always human-gated** in v1
  (ADR 0043/0028), informed by the review bundle (ADR 0057).

Each root (and, later, composite) **owns an integration target**, realized as a git
branch. The landing *levels* are durable; the git *realization* below is process and
evolves by phase.

## Decision — v1 Realization (pre-runtime, single working tree)

- **One integration branch per root**, named `gw/root/<id>-<slug>`. Composite sub-branches
  are deferred until parallel runs exist (Phase 6).
- **`gw` owns the branch lifecycle** — it automates the git the operator did by hand in
  Phase 4, each step still behind its existing gate:
  - the branch is created when the root's **approval envelope** is approved (ADR 0054);
  - a child **`land_to_parent`** is a commit onto the root branch — ADR 0034's commit path,
    retargeted from `main` to the integration target (in a single working tree, that target
    is the checked-out branch, so this is exactly today's behavior);
  - **root `land_to_main`** merges the root branch into `main` on the approved landing,
    then deletes the branch.
- **Merge style: `--no-ff` by default**, configurable (preserves a visible phase/feature
  boundary, as used merging Phase 4).
- **No worktrees in v1.** A single working tree with one active operator is sufficient
  pre-runtime; worktrees buy *parallel isolation*, which only matters once the runtime
  dispatches concurrent children.

This is precisely the workflow already dogfooded by hand (the `phase-4-operator-ui` branch
was the root integration target; per-ticket commits were child landings; the reviewed
`--no-ff` merge to `main` was the root landing). v1 moves that lifecycle into `gw` so it is
tracked, gated, and evidence-backed (ADR 0057) rather than manual.

## Phase 6 Trajectory (not built here)

The same contract scales without changing the envelope/bundle ADRs:

- isolated **worktree per child run** (`gw/run/<id>`) from the parent integration base;
- child results as **result refs/patches** (distinct from ephemeral checkpoints, ADR 0045);
- parent integration by merge/rebase/cherry-pick/patch, with a parent-integration node
  owning conflicts and scope reconciliation (ADR 0045/0046);
- squash-at-landing where desired (existing Phase 6 ticket T-0504).

## Consequences

- The coordinator gains root-branch lifecycle management (create on envelope approval,
  commit child landings to the target, merge on approved root landing, cleanup) and a
  `land_to_parent` path distinct from `land_to_main`.
- ADR 0034 generalizes: landing commits to the node's **integration target**, which in v1
  single-tree mode is the checked-out root branch.
- Conflicts in v1 surface to the human (single linear branch); concurrent-conflict
  handling is a Phase 6 concern.

## Open Questions

- When do composite sub-integration branches become worth it — only with Phase 6
  parallelism, or earlier for very large features? v1: root-only.
- Should the root branch be created at envelope approval or lazily at first child landing?
  v1: at envelope approval, so the target exists before any `land_to_parent`.
- Branch garbage collection and naming collisions across re-planned roots — needs a small
  convention (v1: delete on successful root landing; superseded roots keep their branch
  until resolved).
