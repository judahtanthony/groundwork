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
- **T-0501** runtime interface + Codex adapter shell (ADR 0027) — new `runtime.Codex`
  adapter behind the existing `Runtime` seam: `Config` (command/model/sandbox/args), a
  `LaunchFunc` seam (the real process launcher slots in at T-0502 via `WithLauncher`), and
  a records-only default launch so dispatch stays functional now. `runtime.Select` chooses
  stub vs codex from `config.runtime` (unknown name errors loudly). Added a `model` project
  config key; wired `gw server` boot to select the runtime from config instead of hardcoding
  the stub. `Run` resolves the effective model from the coordinator's actor-selected Spec,
  falling back to config. Tests: selector mapping + unknown, default command, launcher
  delegation with actor config, model fallback, records-only shell sequence.
- **T-0506** per-run git worktree primitives + lifecycle (ADR 0059) — `internal/git`
  gains `WorktreeAdd`/`WorktreeRemove`/`WorktreeList`/`WorktreePrune`, `MergeSquash`,
  `DiffNameOnly`/`DiffRange` (three-dot base diff), `UpdateRef`/`DeleteRef` (the
  `refs/groundwork/runs/<id>` retention namespace), `BranchExists`, and
  `DeleteBranchForce`. New `internal/worktree.Manager` provisions `gw/run/<run-id>` from a
  base commit under `.groundwork/worktrees/<run-id>` (`Project.WorktreesDir()`), retains the
  WIP chain under the run ref before teardown so nothing is dropped, and `Reconcile`s
  orphaned worktrees against the live run set. Tests: add/list/remove, prune orphan,
  base-diff isolates the run's own change, squash stages into the index, ref retention,
  and Manager provision/retain/teardown/reconcile (confirming `.groundwork/worktrees/`
  placement works inside the repo).
- **T-0502** launch Codex in isolated worktree (ADR 0059) — `execLauncher` runs the
  configured agent command with cwd = the run worktree, after `validateWorkspace` confirms
  the cwd is a real directory contained within the worktree root (so the agent can never
  run in the repo root or an arbitrary path); stdout/stderr stream as `output`/`stderr`
  events and the exit code maps to produced/failed/interrupted. The Codex adapter gains
  `WithExec()` (real launcher) and `WithWorkspace(provider, base)` — `Run` provisions the
  run's worktree from the resolved base and sets the workspace before launch; the worktree
  and its branch persist past the run for diff capture and landing. Wired `gw server` to
  give the Codex adapter the exec launcher + a `worktree.Manager`-backed provider when the
  project is a git work tree (else records-only). Tests: agent runs in the workspace and
  streams output, non-zero exit errors, containment validation (empty/outside-root), and
  provision-before-launch delegation.
- **T-0503** stream runtime events to store + JSONL (ADR 0027) — the scheduler's run-event
  sink now mirrors each event to a per-run `events.ndjson` (`appendRunEventLog`, canonical
  one-line-per-event under `.groundwork/runs/<run-id>/`, tier-1 ignored) alongside the
  existing SQLite `run_events` projection; a log failure publishes `run.error` but never
  fails the run. Added `RunLogDir` to scheduler config and wired `gw server` to pass
  `RunsDir()` and the configured model. Run records already carry actor_id/runtime/model
  (ADR 0027). Test: a dispatched run persists events to SQLite and events.ndjson and the
  run record carries actor/runtime/model.
- **T-0507** capture run changed-file set from worktree (ADR 0059) — `worktree.Manager.Diff`
  stages the run worktree and diffs its index against the base (`git diff --cached` via new
  `git.DiffCachedNameOnly`/`DiffCached`), so the changed-file set is captured whether or not
  checkpoints are committed. `runtime.Result` gains `ChangedFiles` + `Diff`; the Codex
  adapter captures them after a successful launch and emits a `diff` event. The scheduler
  records the changed-file set on the run (`runs.changed_files_json`, migration 0009) and
  writes the full diff to `.groundwork/runs/<id>/diff.patch`. `ChangedFilesForNode` exposes
  the latest run's diff for gate inputs (consumed by T-1091); the shared-index land-preview
  path is untouched. Tests: worktree diff captures add/modify/delete and empty runs, store
  round-trip + latest-non-empty selection, and adapter capture into the Result.
- **T-0504** run checkpoints + landing squash (ADR 0015/0059) — `worktree.Manager.Checkpoint`
  commits the run worktree's work as a WIP commit on its `gw/run/<id>` branch (never on
  main/integration); the Codex adapter checkpoints after capturing the diff. `LandToParent`
  now squashes the node's run branch into the checked-out integration branch
  (`squashRunBranch`: `git merge --squash` with a hard-reset recovery on conflict), the
  existing commit path makes the single curated landing commit (squashed code + sidecar),
  then `cleanupLandedRun` retains the WIP chain under `refs/groundwork/runs/<id>` and tears
  down the run branch + worktree. Manual single-tree landings (no run branch) are unchanged.
  Added `git.ResetHard` and `LatestRunIDForNode`. Tests: checkpoint commits + empty no-op,
  end-to-end checkpoint→squash→retain→teardown, and a server `LandToParent` that squashes a
  run branch into the integration branch and reaps it. Completes E-0006; ADR 0059 →
  Implemented.
- **T-1055** blocked-run handoff outcomes + resume packets (ADR 0051) — `runtime.Result`
  gains the outcome vocabulary (produced/completed/blocked/input_required/escalated/rework/
  cancelled/interrupted) plus `HandoffSummary`/`Statement`; `IsBlockedOutcome` drives the
  scheduler: a blocked/escalated/input-required run releases its lease, writes a durable
  blocker via `RecordBlockedHandoff` (input → input_requested, else decision_requested,
  carrying the handoff summary), and moves the ticket to **blocked** (not review). New
  `internal/resume.Assemble` builds the ADR 0051 resume packet from durable state — ticket
  context, nearest ancestor contract, acceptance, dependency statuses, pending blockers vs
  resolved decisions, handoff summary, captured changed files, validation state, and a
  recommended next action. Exposed via `GET /tickets/{id}/resume`, client, and CLI
  (`gw ticket resume`). Tests: packet assembly (blocked + clean), and scheduler blocked
  routing (blocked status, durable handoff, lease released).
- **T-1058** completion + blocked-run handoff summaries (ADR 0047) — the scheduler
  auto-writes a node's completion summary (sidecar + SQLite mirror) from a produced run's
  evidence (changed files, validation, outcome) when none exists, so runtime-produced
  results always carry the summary review/landing requires; a pre-existing human summary is
  left untouched. Reordered the blocked path so the durable handoff summary is written
  before the lease is released. `completion.Stale` detects staleness (node returned to
  rework, or the changed-file set diverged from the summary); `Server.SummaryStale` surfaces
  it for review/context. Added `TicketsDir` to scheduler config (wired in boot). Tests:
  staleness cases, auto-summary on produced + mirror, and existing-summary preservation.
- **T-1056** extend context briefs with durable decisions + summaries (ADR 0051/0047) —
  `contextbrief.Build` now carries pending blockers, recent resolved decisions (bounded to
  the newest 5), the captured changed-file set, the completion summary, and a
  staleness/missing signal (stale via `completion.Stale`; missing flagged for review/done
  nodes). The brief already excludes raw transcripts, so it prefers these summaries + canon
  (contract/SOPs) per the ADR. `gw context` renders the new sections (pending blockers,
  completion summary, ⚠ stale/missing cues). Tests: durable memory in the brief with a
  stale-summary signal, and missing-summary detection at review.
- **T-1090** route scheduler AI claims through envelope-aware authz (ADR 0056) — the
  scheduler gains an optional `EnvelopeGate` seam (`ClaimDecision`: allow/deny/exception);
  `selectActor` routes AI candidates through it instead of the trust-only
  `policies.AuthorizeClaim` when installed — allow dispatches, deny skips, and a boundary
  crossing publishes `claim.exception` without dispatching (the human exception approval was
  already opened by `AuthorizeEnvelopedClaim`). The gate stays free of a server import; the
  CLI wires an `envelopeGate` adapter over the coordinator's new `Server.AuthorizeAIClaim`
  (computes claim-time risk class, runs the AND-composition with no diff yet). Default-deny
  for no-envelope nodes is preserved. Tests: gate allow dispatches even when trust-only would
  not, deny blocks, and exception does not dispatch.
- **T-1091** enforce envelope file-scope + escalation triggers via diff (ADR 0056) — new
  `enforceEnvelopeOnDiff` reads the node's captured changed-file set and the active
  envelope, then enforces actual-vs-planned file scope (`envelopeScopeAllows` for real on a
  non-empty set), `require_review` paths, and the five escalation triggers
  (`on_unexpected_files`, `on_contract_change`, `on_public_api_change`,
  `on_validation_failure`, `on_risk_above_ceiling` — risk computed from the actual diff
  scope). Any breach opens one human exception approval listing the reasons and returns
  `ErrEnvelopeEscalation`; wired into `LandToParent` after the validation gate (mapped to a
  `409 envelope_escalation`), so a breach blocks the landing — the human-gated invariants
  are never bypassed. Ungoverned/manual nodes are a no-op. Tests: helper detection, within-
  scope no-op, each trigger escalates with an exception, and no-envelope no-op. Completes
  E-0012.
