# ADR 0047: Hierarchical Memory And Context Compression

Status: Accepted
Implemented: Partial

## Implementation State

Groundwork already has canon, bounded context briefs, run records, and context-miss capture. This ADR accepts the hierarchical summary direction as the memory model for Phase 4+, but completion summaries, blocked-run handoff summaries, parent memory, dependency-output compression, and stale-summary invalidation remain future implementation work.

## Goal

Make agent context reflect the work hierarchy. A child should receive the context it needs
to satisfy its parent contract without wading through irrelevant sibling transcripts or
early discovery details.

## Problem

As Groundwork decomposes work into many nodes, raw context can become too large and too
noisy. Neighboring child runs may contain false starts, local implementation detail, or
obsolete discoveries. Passing all of that to every later agent increases cost and lowers
focus.

## Decision

Use hierarchical memory compression:

```text
child transcript -> child completion summary -> parent memory -> root summary -> canon
```

Full transcripts and run events remain available for audit and debugging. Normal planning
and execution should use compact summaries, contracts, dependency outputs, and canon.

## Child Context Inputs

A child run should receive:

- global project instructions,
- workflow and relevant SOP,
- root goal summary,
- ancestor contracts,
- approved envelope,
- its own node brief,
- direct dependency summaries,
- resource/file scope,
- relevant validation requirements,
- policy constraints and escalation triggers.

A child should not normally receive:

- every sibling transcript,
- unrelated discovery notes,
- raw logs from completed children,
- obsolete planner hypotheses,
- broad docs beyond the files relevant to the node.

## Completion Summary

Every child that produces a result should emit a compact completion record before review
or landing:

```yaml
completion_summary:
  node_id: T-2001
  outcome: Implemented Codex runtime launch interface.
  changed:
    - internal/runtime/runtime.go
    - internal/runtime/codex.go
  validation:
    - command: go test ./...
      status: passed
  decisions:
    - Codex adapter accepts a workspace path and event sink from the coordinator.
  assumptions:
    - CLI app-server invocation details remain behind adapter config.
  risks:
    - Resume semantics need a follow-up integration test.
  canon_updates:
    - docs/architecture/agent-runtime.md
```

## Parent Memory

A parent node should maintain a synthesized view of child outputs:

- child outcomes,
- integrated result refs,
- unresolved risks,
- interface decisions,
- conflicts or exceptions,
- validation state,
- canon updates needed.

The parent memory should be regenerated or reconciled as children land into the parent.
It should be the default context source for parent integration runs.

Parent memory is durable operational state while the subtree is active. It may be stored
as ticket-attached sidecar records or as a deterministic generated summary derived from
child completion records. Ratified conclusions that affect future work belong in canon.

## Dependency Context

Direct dependencies are more relevant than siblings. When node C depends on node B, C
should receive B's completion summary and any declared interface contract. Other siblings
should be omitted unless the parent integration memory says they matter.

## Canon Distillation

Canon should receive durable conclusions, not raw run notes. A node can propose canon
updates, but accepted ADRs, architecture docs, contracts, policies, and SOPs remain the
durable memory for future sessions.

## Blocked-Run Handoff

A blocked autonomous run should emit a compact handoff summary before it exits. The
summary explains the blocker, current state, useful artifacts/checkpoints, attempted
options, recommendation, and next action. If the blocker must survive rebuild, the
handoff is referenced from a durable ticket decision record (ADR 0051).

## Staleness

Summaries are stale when the underlying child result, checkpoint/diff, parent contract,
dependency output, or rework decision changes. Context assembly must prefer the latest
non-superseded summary and surface stale or missing summaries as recovery/rework
signals rather than silently using obsolete context.

## Web Admin Implications

The web admin should let reviewers inspect three layers:

- concise child summary,
- expanded transcript/logs for audit,
- parent synthesized memory and unresolved risks.

This supports efficient review without hiding detail.

## Consequences

Completion summaries and blocked-run handoff summaries become required runtime outputs
for Phase 4 work that reaches review/landing or exits blocked. Context assembly should
prefer summaries, contracts, dependency outputs, and canon, adding transcript excerpts
only when they are directly relevant. Parent memory and stale-summary invalidation need
implementation tickets before the model is complete.
