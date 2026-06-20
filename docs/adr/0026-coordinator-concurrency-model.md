# ADR 0026: Coordinator Concurrency Model

Status: Accepted

## Context

`gw server` owns multi-agent scheduling (`coordinator.md`). Phase 1 already
provides the load-bearing concurrency primitive: `DB.ClaimTicket` is a single
transaction that verifies eligibility, rejects a live lease, records the lease,
and moves the node to `in_progress` so only one run can win
(`internal/store/sqlite/lease.go`). Phase 2 must decide how the coordinator
*drives* that primitive concurrently — a single event loop versus a
goroutine-per-run free-for-all — and how leases are kept alive, without
reintroducing the split-brain that `coordinator.md` warns against.

## Decision

A **single scheduler event loop** plus a **bounded pool of run supervisors**.

- **Serial scheduler tick.** One goroutine runs the scheduler. Each tick it asks
  the store for eligible nodes (`ListEligible`, which already recomputes
  dependency satisfaction per call), applies actor selection and the gate engine
  ([ADR 0028](0028-gate-evaluation-engine.md)), and calls `ClaimTicket` for the
  selected node. Claims are therefore serialized through one goroutine *and* a DB
  transaction; the transaction remains the correctness boundary even if the loop
  is ever parallelized later.
- **Bounded run supervisors.** Each successful claim starts one run-supervisor
  goroutine, up to `config.MaxConcurrency` (default 4). The scheduler does not
  launch new claims while at capacity. A supervisor owns one run's lifecycle
  ([ADR 0027](0027-run-lifecycle-and-checkpoint-records.md)) and renews its lease
  on the **heartbeat interval** (default 30s) against the **TTL** (default 90s),
  both from `config.Lease`. If a supervisor dies, its lease expires and the node
  becomes reclaimable on a later tick (and is reconciled at startup per
  `recovery.md`).
- **In-process event hub.** A small pub/sub hub fans run/state/approval/
  validation events to SSE subscribers ([ADR 0025](0025-http-server-and-sse-transport.md))
  and persists them to `run_events`. The hub is in-memory, bounded, and
  non-blocking: a slow subscriber is dropped, not allowed to stall the loop.
- **Shutdown** stops admitting claims, asks supervisors to pause at a safe
  boundary, flushes events, and marks anything still active `interrupted`
  (`recovery.md`).

## Parallelism Is Preserved

The serial element is **claim arbitration only**, not execution. Two layers run
concurrently:

- **Execution fans out.** Each claim hands off to its own supervisor goroutine,
  so up to `MaxConcurrency` runs execute **in parallel**, and they may be
  **different, specialized actors** (the loop selects an actor per node via the
  gate engine, [ADR 0028](0028-gate-evaluation-engine.md), [ADR 0029](0029-actor-identity-model.md)).
  Width is a config dial, not a fixed `1`.
- **The DAG is what makes fan-out safe.** The dependency overlay
  ([ADR 0010](0010-dependencies-and-upward-revision.md)) guarantees that every
  node in the eligible set at a given tick is independent of the others, so
  dispatching them concurrently to separate agents cannot race on shared work.
  Strong dependencies are precisely the enabler of reliable multi-agent fan-out,
  not a constraint on it.

Serializing only the *claim transaction* (which must be single-winner regardless)
keeps correctness simple while leaving execution width fully parallel. Future
scaling — sharded loops or parallel claim workers — remains open because the
claim is already transactional.

## Consequences

`internal/scheduler` (the loop + selection) and the run supervisor live in the
coordinator; `internal/server` only triggers and observes them. The model is
deliberately conservative: contention is resolved by SQLite transactions, not by
in-process locking gymnastics, which keeps the concurrency story testable with
`-race` and consistent with the "active orchestration goes through the
coordinator" rule. Raising throughput later (sharded loops, parallel claims) is
possible because the transactional claim already tolerates concurrency.
