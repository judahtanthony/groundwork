# Recovery

Groundwork must relaunch from a stable point after crash or shutdown.

## Startup

`gw server` startup should:

1. Open or create `.groundwork/state.sqlite`.
2. Run schema migrations.
3. Load config and policy files.
4. Import committed ticket exports when SQLite is missing or explicitly requested, rebuilding the work-node hierarchy and dependency edges.
5. Import ticket sidecar decision records, including pending approvals, input requests,
   decision requests, rework requests, and recovery records.
6. Recreate pending approval/input/decision queues from durable pending records.
7. Verify node rows have exported records and that dependency edges remain acyclic.
8. Detect inconsistent states: blocked tickets with no dependency or durable blocker,
   review tickets with no recoverable proposal/gate record, and runtime-only approvals
   whose payload was required but not exported.
9. For unrecoverable lost context, create or surface `recovery_needed` rather than
   silently stranding the ticket.
10. Mark stale `running` runs as `interrupted` unless a live worker process is verified.
11. Expire leases whose worker is gone or whose TTL elapsed.
12. Inspect run directories and attach recoverable logs to interrupted runs.
13. Verify worktree paths are inside `.groundwork/worktrees/`.
14. Regenerate views from SQLite.
15. Resume scheduling.

## Clean Teardown

Clean shutdown should:

1. Stop accepting new runs.
2. Ask active runs to pause at a safe boundary.
3. Flush run events and transcripts.
4. Export ticket state and timelines.
5. Mark remaining active runs `interrupted` if they cannot stop cleanly.
6. Release or expire leases according to shutdown mode.

## Crash Recovery

Completed and landed work is preserved through Git commits and ticket exports. Active runs may lose model-internal context, but node state, durable ticket-attached request records, run manifest, transcript, and the worktree's last **checkpoint** commit should be sufficient to resume with a new agent turn — uncommitted in-flight work is preserved by run checkpoints rather than lost (see [ADR 0015](../adr/0015-run-checkpoints-squashed-at-landing.md)). Per-node journals are ephemeral; durable design is already in canon because it is distilled at the ratification gate, not at crash time.

Recovery cannot reconstruct blocker payloads that were never exported. When only a lost
runtime handle remains, Groundwork must choose an explicit recovery path: rerun
planning/action, move the ticket to rework, create a recovery ticket, ask a human to
decide manually, or recreate a live approval from durable state when safe.
