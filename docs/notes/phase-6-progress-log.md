# Phase 6 — Durable Handoff & Codex Runtime: implementation progress log

Autonomous execution of the Phase 6 leaf chain (ADRs 0051/0053/0027/0015/0058/0059) on
`phase-6-durable-handoff-codex-runtime`. One entry per landed leaf; build/vet/test green
at each step.

## Decomposition (T-1071)

- Drafted and ratified **ADR 0059** (worktree-per-run topology and land-from-worktree
  integration): `gw/run/<run-id>` from the integration base under
  `.groundwork/worktrees/<run-id>/`; checkpoints squashed into `gw/root/<id>-<slug>` at
  `land_to_parent`; run diff-vs-base becomes the authoritative changed-file set for gates;
  reconciled the ADR 0015 vs 0058/0045 branch/ref naming.
- Triaged the existing T-1071 tree (E-0006 runtime, T-1052 durable handoff, T-0904,
  T-1003) and added the seams Phase 5 left inert:
  - **T-0506** per-run git worktree primitives + lifecycle (E-0006).
  - **T-0507** capture run changed-file set from worktree, feed gate inputs (E-0006).
  - **E-0012** (new epic) activate envelope/escalation enforcement with the runtime diff,
    with **T-1090** (envelope-aware scheduler claim routing) and **T-1091** (file-scope +
    escalation-trigger enforcement against the diff).
- Filled thin descriptions, set the dependency DAG, priorities, and `status: todo` across
  the Phase 6 tree (validated: parents/deps resolve, no cycles).

## Landed leaves

- **T-1053** ticket decision sidecar import/export (ADR 0051/0053) — new
  `internal/decision` package: the durable `Record` type (full decision-records.md
  shape), canonical NDJSON encode/decode (compact, fixed key order, sequence-ordered,
  byte-stable), and `decisions.ndjson` sidecar read/write (optional — removed when no
  records). Added a SQLite projection (`decisions` table, migration 0008;
  `AppendDecision`/`ImportDecision`/`ListDecisions`/`ListPendingDecisions`/`HasDecisions`)
  with the sidecar authoritative. Wired `gw export` to write the sidecar from the store
  and `gw ticket import` to project it back. Tests: encode/decode round-trip + determinism
  + one-line-per-record, validate required fields, unknown-field rejection, sidecar
  write/read/empty-removal, store append/seq/pending/import-preserves-seq, and a CLI
  import→re-export byte-stability round-trip.
- **T-1054** rebuild live queues from durable records (ADR 0051) — new
  `RebuildDurableQueues` (store) projects pending durable records into live queues at
  startup and surfaces stranded tickets: pass 1 recreates an `approvals` row (fresh A-id
  runtime handle) for each pending `approval_requested` record lacking a live row;
  `input_requested`/`decision_requested` records stay the durable explainer for a blocked
  ticket (no extra table). Pass 2 appends a `recovery_needed` record to any
  blocked/review/rework ticket with no durable explainer and no pending approval. Wired
  into `gw server` boot after `ReconcileStartup`; idempotent. Tests (DB purge/rebuild):
  decompose/replan/land_to_main approval recreation + idempotence, input_required stays
  explained, and recovery_needed surfaced for stranded blocked/review tickets.
