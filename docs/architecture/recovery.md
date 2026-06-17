# Recovery

Groundwork must relaunch from a stable point after crash or shutdown.

## Startup

`gw server` startup should:

1. Open or create `.groundwork/state.sqlite`.
2. Run schema migrations.
3. Load config and policy files.
4. Import committed ticket exports when SQLite is missing or explicitly requested, rebuilding the work-node hierarchy and dependency edges.
5. Verify node rows have exported records and that dependency edges remain acyclic.
6. Mark stale `running` runs as `interrupted` unless a live worker process is verified.
7. Expire leases whose worker is gone or whose TTL elapsed.
8. Inspect run directories and attach recoverable logs to interrupted runs.
9. Verify worktree paths are inside `.groundwork/worktrees/`.
10. Regenerate views from SQLite.
11. Resume scheduling.

## Clean Teardown

Clean shutdown should:

1. Stop accepting new runs.
2. Ask active runs to pause at a safe boundary.
3. Flush run events and transcripts.
4. Export ticket state and timelines.
5. Mark remaining active runs `interrupted` if they cannot stop cleanly.
6. Release or expire leases according to shutdown mode.

## Crash Recovery

Completed and landed work is preserved through Git commits and ticket exports. Active runs may lose model-internal context, but node state, run manifest, transcript, and the worktree's last **checkpoint** commit should be sufficient to resume with a new agent turn — uncommitted in-flight work is preserved by run checkpoints rather than lost (see [ADR 0015](../adr/0015-run-checkpoints-squashed-at-landing.md)). Per-node journals are ephemeral; durable design is already in canon because it is distilled at the ratification gate, not at crash time.

