# ADR 0027: Run Lifecycle And Checkpoint Records

Status: Accepted
Implemented: Partial

## Context

`runtime-model.md` defines two run **modes** (planning and implementation) and a
checkpoint mechanism ([ADR 0015](0015-run-checkpoints-squashed-at-landing.md)),
and `sqlite-schema.md` defines the `runs`/`run_events` tables that Phase 1 did
not create. The Codex runtime that actually executes runs — isolated worktrees,
real event streaming, real WIP commits — is Phase 4 (E-0006). Phase 2 still needs
run *records*, lifecycle state, actor snapshots, pause/resume/cancel, and the
checkpoint/landing-squash *semantics* so approvals, scheduling, recovery, and
distillation have something concrete to operate on. The question is how much of
the run subsystem is records-only in M2 versus deferred with the runtime.

## Decision

Model the run lifecycle and checkpoints as **records and state transitions in
M2**, with a **records-only stub runtime**, and defer all real agent/git
execution to Phase 4.

- **Run record + modes.** A run carries `mode` (`planning` | `implementation`),
  `status`, `actor_id`, `actor_snapshot_json`, `runtime`, `model`,
  `workspace_path`, `base_commit`, timestamps, and token counters per
  `sqlite-schema.md`. The actor snapshot is captured from `.groundwork/actors.yaml`
  at run start so audit survives later registry edits (`actors.md`, [ADR 0023](0023-actors-work-types-and-policy-routing.md)).
- **Lifecycle states.** `pending → running → {paused ⇄ running} → completed`,
  with `cancelled` and `interrupted` as the off-ramps. Pause/resume/cancel are
  CLI/API-driven, audited, and (for cancel) release the lease. `interrupted` is
  set only by recovery (`recovery.md`), never by a client.
- **`runtime.Interface` + records-only stub.** A minimal runtime interface is
  introduced now; its only M2 implementation is a **stub that emits synthetic
  lifecycle events and writes no code**. This exercises scheduler → run → events
  → gate → landing → recovery end-to-end without launching Codex, and is the
  exact seam the Codex adapter fills in Phase 4 (E-0006). The stub is justified:
  the coordinator cannot be integration-tested otherwise, and building it as the
  interface (not a throwaway) means no rework.
- **Checkpoints as records, squash as semantics.** M2 stores checkpoint records
  and the landing rule that WIP checkpoints are squashed and never reach `main`
  ([ADR 0015](0015-run-checkpoints-squashed-at-landing.md)). The *actual* git
  commits on the worktree branch, the `refs/groundwork/runs/<run-id>` namespace,
  and resume-from-checkpoint (T-0904) are Phase 4, because they require the real
  worktree the runtime owns. M2's landing flow marks checkpoints squashed and
  enforces the gate; it does not run git.

## Consequences

`runs`/`run_events` migrations land in M2 (T-0410); `internal/run` holds the
status/mode state machines; `internal/runtime` holds the interface + stub; the
run supervisor ([ADR 0026](0026-coordinator-concurrency-model.md)) drives the
lifecycle. Checkpoints are recorded in M2 as `run_events` rows (`checkpoint`,
`checkpoints_squashed`) rather than a dedicated table — the durable checkpoint
artifact is the git ref, which the Phase 4 runtime creates. `recovery.md`'s "mark stale runs interrupted / expire
leases" is fully implementable now; "resume from last checkpoint" is explicitly
Phase 4. The records-vs-execution split keeps the durable-state boundary
([ADR 0012](0012-three-tier-state-and-ratification-timing.md)) intact: run
manifests are durable operational records, worktree contents and transcripts
stay tier-1 ephemeral.
