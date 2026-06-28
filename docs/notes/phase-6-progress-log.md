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
- **T-1057** consequential decision ticket routing (ADR 0052) — new store
  `RaiseDecision` composes existing durable primitives into the consequential branch of
  the ladder: creates a `kind: decision` leaf work node (routed by `work_type`, parent
  inherited), links the blocked ticket to it, appends a durable `decision_requested`
  record pointing at the node, and transitions the blocked ticket to `blocked` — no
  separate decision subsystem. `RequestInput` is the small-uncertainty branch: a durable
  `input_requested` record with no node and no edge. Exposed via API
  (`POST /tickets/{id}/decision|input`), client, and CLI (`gw ticket decision|input`).
  Tests: node/edge/record/blocked-transition created; required work_type+statement; input
  request creates no ticket and no edge and leaves status unchanged.
- **T-1059** file-authoritative durable ticket state (ADR 0053) — store-level filesystem
  write-through, opt-in via `SetExportDir`: `CreateTicket`/`UpdateTicket`/`TransitionTicket`/
  `TriageTicket`/`AddDependency`/`RemoveDependency`/`Reparent`/`AppendDecision` and the
  decompose propose/accept/reject paths now rewrite the affected node's sidecar
  (ticket.md + decisions.ndjson) before reporting success. Pending human-gated approvals
  emit a paired durable `approval_requested` record (correlated by approval id) on
  `Request` and are resolved in place on `Decide`, so they survive rebuild and reproject
  via T-1054 without phantoms. `DetectFileDivergence` compares each ticket's canonical
  export to its committed sidecar at startup and appends a non-destructive `recovery_needed`
  to the node's sidecar rather than silently trusting SQLite. Wired into `gw server` boot
  (import → enable write-through → divergence check → reconcile → queue rebuild) and CLI
  `openStore`; import disables write-through for its own duration. Tests: purge/rebuild
  preserves tickets/statuses/deps/decisions; divergence flags an unexported mutation
  (idempotently); pending approval emits + resolves its durable record. ADR 0051/0053 →
  Implemented: Partial.
