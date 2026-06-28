# ADR 0059: Worktree-Per-Run Topology And Land-From-Worktree Integration

Status: Accepted
Implemented: Implemented

Concretizes the Phase 6 worktree trajectory sketched by
[ADR 0045](0045-node-branching-and-parent-integration.md) (node branching) and the
"Phase 6 Trajectory" of [ADR 0058](0058-integration-targets-and-landing-levels-v1.md)
into a buildable v1 mechanism, and reconciles the branch/ref naming left inconsistent
between [ADR 0015](0015-run-checkpoints-squashed-at-landing.md) and ADR 0045/0058. It is
the git substrate the Phase 6 runtime (E-0006) and the diff-fed gate activation (E-0007)
depend on.

## Context

The Phase 5 landing model (ADR 0058) runs in a **single working tree**: a child
`land_to_parent` is just a commit onto the checked-out root integration branch
`gw/root/<id>-<slug>`, and the changed-file set that feeds gates comes from that one
tree's git index (`s.repo.StagedDiff()`). Phase 6 introduces **concurrent** runs, so each
run needs an isolated working tree; the single-index assumptions break.

Three concrete decisions were left open by prior ADRs and block implementation of T-0506
(worktree provisioning), T-0502 (launch), T-0504 (checkpoints/squash), and T-0507 (diff
capture):

1. **Where worktrees live on disk**, how they stay out of the indexed source tree and
   `git status`, and how orphans are reclaimed.
2. **How `land_to_parent` integrates a run's worktree branch** into the root integration
   branch — merge, rebase, cherry-pick, or patch — given ADR 0015 requires WIP checkpoints
   to be squashed and never reach durable history.
3. **Branch/ref naming**: ADR 0015 illustrates `agent/<node-id>` plus a throwaway
   `refs/groundwork/runs/<run-id>` namespace; ADR 0058/0045 use `gw/run/<id>`. ADR 0045
   additionally leaves "git refs, patch files, or both?" unresolved.

## Decision

### Worktree per run

Each dispatched run executes in its own git worktree created with `git worktree add`,
checked out on a fresh run branch **`gw/run/<run-id>`** branched from the run's **base
commit** — the tip of the node's integration target (the root branch `gw/root/<id>-<slug>`
when one exists, else the repo default branch). The base commit is recorded on the run
(`runs.base_commit`, ADR 0027) so the diff and integration are computed against a fixed
point even as the integration branch advances.

**Location.** Worktrees live under `.groundwork/worktrees/<run-id>/`, consistent with the
project's "runtime artifacts under `.groundwork/`" model (ADR 0003/0007). `.groundwork/`
runtime state is already git-ignored and excluded from indexing; `worktrees/` is added to
that exclusion so nested worktrees never pollute the root `git status` or the CodeGraph
index. The path is recorded on the run record (`runs.workspace_path`). A configurable
override is allowed but defaults here so discovery and cleanup are deterministic.

**Lifecycle.** A worktree is provisioned at run start and removed at run end
(`git worktree remove`), after a successful `land_to_parent` or a terminal
cancel/interrupt. Recovery enumerates live worktrees with `git worktree list` and
reconciles them against run records: a worktree with no live run is pruned; a run whose
worktree is gone is `interrupted`. `git worktree prune` clears stale administrative
metadata. Cleanup never deletes a worktree with un-landed, un-checkpointed changes
without surfacing a `recovery_needed` signal (ADR 0051).

### Checkpoints, results, and land-from-worktree

**Result representation is a git branch, not a patch file** (resolving ADR 0045's open
question toward branches for v1; patches/cherry-pick remain a later option for cross-base
integration). The run's WIP **checkpoints are ordinary commits on `gw/run/<run-id>`**
(ADR 0015) inside the run worktree, never on `main` or the integration branch.

`land_to_parent` integrates the run branch into the root integration branch by
**squash** (ADR 0015): the net diff of `gw/run/<run-id>` against its base is collapsed into
a **single curated commit** on `gw/root/<id>-<slug>`. In v1 the integration branch stays
checked out in the **main** working tree (ADR 0058), so the squash is realized there via
`git merge --squash gw/run/<run-id>` followed by the existing ADR 0034 commit path,
retargeted from `main` to the integration branch. The WIP checkpoint chain does not reach
the integration branch; it is retained only under the throwaway ref namespace
**`refs/groundwork/runs/<run-id>`** (ADR 0015) for post-mortem/recovery and is garbage-
collectable. `root → main` landing is unchanged (ADR 0058: human-gated `--no-ff` merge).

Conflicts surface to the human, as in ADR 0058 — v1 integrates onto one linear root
branch and does not auto-resolve concurrent overlapping edits; the scheduler avoids
dispatching parallel children with overlapping exclusive scope (ADR 0046).

### Diff source of truth

The changed-file set that feeds gates — validation-template selection, envelope file-scope
(`envelopeScopeAllows`), and the escalation triggers (E-0007) — is the **run worktree's
diff against its base commit** (`git diff --name-only <base>...gw/run/<run-id>`), captured
on the run (T-0507). This replaces the shared-index `StagedDiff()` as the authoritative
diff for runtime-produced work. The land-preview path (ADR 0058 / T-1073) continues to
show the staged change set of whichever tree is landing.

### Naming reconciliation

`gw/run/<run-id>` is the canonical run-branch name (supersedes ADR 0015's illustrative
`agent/<node-id>`). `refs/groundwork/runs/<run-id>` is retained as the throwaway
pre-squash checkpoint namespace. `gw/root/<id>-<slug>` remains the per-root integration
branch (ADR 0058). This ADR refines, and does not relax, ADR 0015 (squash-at-landing) or
ADR 0058 (landing levels and human-gated root landing).

## Consequences

- `internal/git` gains worktree primitives: `WorktreeAdd`, `WorktreeRemove`,
  `WorktreeList`/prune, `MergeSquash`, base-diff helpers, and the
  `refs/groundwork/runs/<run-id>` ref handling. (T-0506.)
- The scheduler/runtime provision a worktree per run and tear it down on completion
  (T-0502); checkpoints commit on the run branch and land squashes into the integration
  branch (T-0504); the run's base-diff is captured and exposed to gate inputs (T-0507).
- `.groundwork/worktrees/` is added to git-ignore and indexing exclusions.
- `recovery.md` gains worktree reconciliation via `git worktree list` against run records.
- `land_to_parent` keeps its v1 single-tree integration-branch commit path; the only
  change is that the committed content is produced by squashing a run worktree branch
  rather than committing the operator's manual edits.
- The durable-state boundary (ADR 0012/0007/0053) holds: worktree contents and the WIP
  checkpoint chain are tier-1 ephemeral; the durable artifacts are the squashed integration
  commit, run records, and ticket-attached handoff state.
