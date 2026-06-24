# Proposal 0003: Node Branching And Parent Integration

Status: Draft

## Goal

Align execution, landing, and review with the work hierarchy. Children should be able to
execute independently and integrate upward, while humans review feature-level outcomes at
root or parent boundaries.

## Problem

If every child lands directly to `main`, small-node decomposition creates many human
landing approvals and increases the risk that parallel work disrupts other roots. If all
children work directly on one branch, agents can step on each other and make recovery
harder.

## Proposal

Each root or composite node may own an integration target. Child nodes execute in
isolated worktrees from a known parent commit. Completed children produce a result branch,
patch, or checkpoint set. The parent integrates child outputs, resolves conflicts, runs
broader validation, and summarizes the result.

Root landing to the project branch remains separately gated.

## Conceptual Branch Shape

```text
main
  -> gw/root/T-2000-feature-x
       -> gw/run/T-2001-runtime-adapter
       -> gw/run/T-2002-tests
       -> gw/run/T-2003-docs
```

Parallel children can start from the same parent base:

```text
A = parent integration base
B = child 1 result from A
C = child 2 result from A
D = parent integration of B + C
```

The parent integration step may merge, rebase, cherry-pick, or apply patches. The durable
contract should not require one git operation forever; it should require that the parent
records how child outputs were integrated and validated.

## Landing Levels

- **Child landing to parent**: integrates a child result into its parent integration
  target. This can be auto-approved or batched when it stays inside the approved envelope.
- **Parent landing to parent**: integrates a completed composite into its own parent.
- **Root landing to main**: final feature-level review and commit or merge into the
  project branch. This remains human-gated in v1.

## Parent Integration Node

A parent may need an explicit integration child when sibling work is non-trivial:

- collect child summaries,
- apply child result branches or patches,
- resolve file conflicts,
- reconcile API or behavior mismatches,
- run parent-level validation,
- update docs/canon,
- prepare final review.

This lets narrow child agents stay focused while the parent owns cross-cutting consistency.

## State Records

Each run should record:

- base commit,
- workspace path,
- result ref or patch id,
- files touched,
- validation results,
- summary,
- parent integration target,
- whether result was integrated,
- integration commit or merge record.

Candidate result metadata:

```yaml
node_result:
  node_id: T-2001
  run_id: R-123
  base_commit: abc123
  result_ref: refs/groundwork/runs/R-123
  touched_files:
    - internal/runtime/runtime.go
    - internal/runtime/codex.go
  validation:
    status: passed
    commands:
      - go test ./...
  integrated_into:
    node_id: T-2000
    commit: def456
```

## Conflict Handling

The scheduler should avoid dispatching parallel children with overlapping exclusive
resource scopes. If conflicts appear anyway, the parent integration node owns resolution.
Unexpected conflicts should raise an exception approval if they imply scope expansion or
parent contract change.

## Relationship To Checkpoints

Run checkpoints remain ephemeral WIP recovery points. A child result is a deliberate
output selected for parent integration. The system should keep this distinction clear:

- checkpoints are for resume and recovery,
- result refs or patches are for integration and review.

## Web Admin Implications

For each root, the web admin should show:

- current integration branch,
- child result status,
- conflicts,
- validation state,
- aggregate diff,
- child summaries,
- final root review gate.

## Open Questions

- Should root integration branches be mandatory or introduced only when parallel children
  exist?
- Should child results be represented as git refs, patch files, or both?
- How should root branches be named and garbage-collected?
- Should parent integration be performed by a specialized integration agent?

## Candidate ADRs

- Parent integration branches for root and composite nodes.
- Child result refs as distinct from run checkpoints.
- Root landing as final feature-level review.

