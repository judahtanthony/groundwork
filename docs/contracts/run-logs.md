# Run Logs Contract

Run logs are local runtime records ignored by default.

Expected directory:

```text
.groundwork/runs/<run-id>/
  manifest.yaml
  events.ndjson
  transcript.md
  approvals/
  diffs/
  artifacts/
  checkpoints/
```

## Events

`events.ndjson` is append-only structured telemetry for run lifecycle, actor messages, tool calls, approvals, validation, and state changes. For planning runs it also records triage, decomposition proposals, and escalation events. The run manifest records the run mode (planning or implementation), actor id, runtime/model metadata, and an actor snapshot summary.

## Transcript

`transcript.md` is a human-readable projection for inspection and possible excerpting
into a resume packet. It may contain sensitive data and is not committed by default.
It is evidence, not authoritative memory.

## Checkpoints

`checkpoints/` records the run's work-in-progress commit points. A run periodically commits WIP on its worktree branch so recovery can resume from the last checkpoint; these commits are squashed into the landing commit and never reach `main` (see [ADR 0015](../adr/0015-run-checkpoints-squashed-at-landing.md)). Checkpoints may be stored under a throwaway ref namespace (`refs/groundwork/runs/<run-id>`).

## Journal And Distillation

A per-node **journal** of in-progress decision notes is ephemeral (ignored by default), one file per node so parallel runs never conflict. At the ratification gate, durable design is distilled from the journal into canon (committed docs/ADRs/policy/SOPs); the journal itself is not committed (see [ADR 0013](../adr/0013-canon-as-memory.md)).

Blocked-run handoff summaries and request payloads that must survive rebuild do not
belong only in the run log. They are exported as ticket-attached decision records under
`.groundwork/tickets/<id>/decisions.ndjson` (see [decision-records.md](decision-records.md)).

## Summaries

Runs should produce compact summaries for normal execution context:

- `completion_summary` before review/landing when a node produces a result,
- `handoff_summary` before a blocked run exits,
- validation and checkpoint/diff refs needed by a later run.

Full transcripts remain available for audit; context assembly should prefer summaries,
contracts, dependency outputs, and canon.

## Artifacts

Raw command output, screenshots, and large logs belong under `artifacts/` and are ignored by default.
