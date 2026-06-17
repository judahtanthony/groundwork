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

`events.ndjson` is append-only structured telemetry for run lifecycle, agent messages, tool calls, approvals, validation, and state changes. For planning runs it also records triage, decomposition proposals, and escalation events. The run manifest records the run mode (planning or implementation).

## Transcript

`transcript.md` is a human-readable projection for inspection and resume context. It may contain sensitive data and is not committed by default.

## Checkpoints

`checkpoints/` records the run's work-in-progress commit points. A run periodically commits WIP on its worktree branch so recovery can resume from the last checkpoint; these commits are squashed into the landing commit and never reach `main` (see [ADR 0015](../adr/0015-run-checkpoints-squashed-at-landing.md)). Checkpoints may be stored under a throwaway ref namespace (`refs/groundwork/runs/<run-id>`).

## Journal And Distillation

A per-node **journal** of in-progress decision notes is ephemeral (ignored by default), one file per node so parallel runs never conflict. At the ratification gate, durable design is distilled from the journal into canon (committed docs/ADRs/policy/SOPs); the journal itself is not committed (see [ADR 0013](../adr/0013-canon-as-memory.md)).

## Artifacts

Raw command output, screenshots, and large logs belong under `artifacts/` and are ignored by default.

