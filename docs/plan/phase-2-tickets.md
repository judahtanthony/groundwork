# Phase 2 (M2 — Coordinator) Ticket Breakdown

> **Status: complete.** All six waves are implemented, tested (unit + `-race` on
> concurrency paths), and exercised end-to-end by `scripts/smoke.sh`. The Codex
> runtime is a records-only stub (Phase 4). See ADRs 0025–0031.


Dependency-ordered breakdown for **M2: Coordinator Foundation** (see
`docs/product/roadmap.md`, `docs/plan/milestones.md`). Ticket ids continue the
`work-tree.yaml` convention; new tickets fill gaps within their epic's id range
and are marked **(new)**. Each ticket cross-references its `work-tree.yaml`
epic/ticket and the governing ADR(s).

## Scope boundary (decided)

**In M2:** `gw server` + HTTP API, dependency- and actor-aware scheduler over the
existing transactional claim, run records + lifecycle (records-only, no Codex),
actor snapshots, decomposition proposals, escalation/re-plan routing, approval
records with reversibility + actor gating, validation templates + results +
landing gate, canon distillation plumbing + parent-contract channel, SSE stream,
startup reconciliation, cold-start import.

**Deferred:**
- Dashboard HTML shell + board/detail pages (**T-0801, T-0802**) → later
  web-surface phase. SSE/API only in M2 (AGENTS.md; `dashboard.md`).
- Codex runtime core — launch, isolated worktree, real event streaming, real git
  checkpoints, resume-from-checkpoint (**T-0501, T-0502, T-0503, T-0504,
  T-0904**) → Phase 4 / M4. M2 carves in a `runtime.Interface` + records-only
  stub ([ADR 0027](../adr/0027-run-lifecycle-and-checkpoint-records.md)).
- Generated-view export (**T-0903**) → ties to dashboard.
- Dogfooding (**E-0011**) → Phase 3 / M3.
- Chat-approval & reviewer-agent adapters, policy-learning suggestion endpoints,
  budget gates, autonomy *elevation* mechanics, external bridges → roadmap "Phase
  2 features beyond v1" / Phase 5. Records/engine designed to admit them later.

## Gates (apply to every ticket)

- `gofmt -l` clean and `go vet ./...` clean.
- `CGO_ENABLED=0 go build ./... && CGO_ENABLED=0 go test ./...` pass.
- Concurrency tickets additionally pass `go test -race ./...`.
- Server tickets extend `scripts/smoke.sh` and keep it green.
- Behavior changes update the relevant contract/architecture doc and ADR.

Below, only the **distinguishing** acceptance criteria and gate per ticket are
listed; the standard gates above always apply.

---

## Wave 0 — Schema & server foundation

### T-0410 (new) — Runs/approvals/validation migration `0003`
- **Epic:** E-0005. **ADR:** schema contract (`sqlite-schema.md`).
- **Dep:** none.
- **Acceptance:** migration `0003` creates `runs`, `run_events`, `approvals`,
  `validation_results` with the columns/indexes in `sqlite-schema.md`; forward-only
  and checksum-stable ([ADR 0018](../adr/0018-embedded-ordered-migrations.md)).
- **Gate:** `go test ./internal/store/...`; re-running `Migrate` is a no-op;
  `AppliedMigrationIDs` includes `0003`.

### T-0401 — `gw server` process + health + `/api/v1/state`
- **Epic:** E-0005 (`T-0401`). **ADR:** 0025, 0026, 0031.
- **Dep:** T-0410.
- **Acceptance:** server binds `config.Server.Addr` (default `127.0.0.1:4500`);
  `/healthz` reports store availability; `GET /api/v1/state` returns active
  counts. Clean shutdown.
- **Gate:** `scripts/smoke.sh` starts the server, gets `200` on health, stops it.

### T-0411 (new) — Read API: tickets/children/context/deps/actors
- **Epic:** E-0005. **ADR:** 0025.
- **Dep:** T-0401.
- **Acceptance:** `GET` endpoints from `http-api.md` for tickets, ticket,
  children, context, dependencies, actors, actor; error envelope on `404`.
- **Gate:** `httptest` table tests covering success + envelope.

### T-0412 (new) — Coordinator client + mutating ticket endpoints
- **Epic:** E-0005. **ADR:** 0031, 0025.
- **Dep:** T-0411.
- **Acceptance:** `POST /tickets`, `PATCH /tickets/:id`,
  `POST /tickets/:id/transition`, dependency add/remove; CLI client prefers API
  when reachable and falls back to direct store only for store-safe commands;
  coordinator-required commands fail with `coordinator_required` when server down.
- **Gate:** `httptest` + CLI fallback test with server up and down.

---

## Wave 1 — Policy, actor identity, gate engine

(Can begin after T-0410; independent of the server wave.)

### T-0601 — Load trust + autonomy + validation policy YAML
- **Epic:** E-0007 (`T-0601`). **ADR:** 0028.
- **Dep:** none.
- **Acceptance:** `internal/policy` parses/validates the three policy schemas in
  `policies.md`; invalid policy yields an actionable error; unknown keys warn.
- **Gate:** parse/validate unit tests incl. malformed inputs.

### T-0613 (new) — Actor identity resolution (tiered + class→instance)
- **Epic:** E-0007. **ADR:** 0029.
- **Dep:** none.
- **Acceptance:** dotted-path **prefix matcher** (`agent` matches
  `agent.codex.default`); `type` validated consistent with id root; capability
  predicate matching as an independent set; `requested_actor` (class/prefix) →
  concrete registry instance resolution (one-instance-per-class for M2).
- **Gate:** prefix-match, type/root-consistency, and resolver unit tests.

### T-0605 — Reversibility classifier
- **Epic:** E-0007 (`T-0605`). **ADR:** 0014, 0028.
- **Dep:** none.
- **Acceptance:** classifies an action's scope (files, commands, external
  effects) reversible vs irreversible; irreversible cases (non-reversible
  migration, external/prod state, destructive command, credential access)
  detected; a rule may assert `reversible: false`.
- **Gate:** unit tests over reversible/irreversible scopes.

### T-0610 (new) — Risk classification
- **Epic:** E-0007. **ADR:** 0028.
- **Dep:** none.
- **Acceptance:** maps scope to a 0–100 score and to a class
  (`low` 0–33 / `medium` 34–66 / `high` 67–100 / forced `critical`); score is
  display/ranking only.
- **Gate:** mapping + boundary unit tests.

### T-0614 (new) — Gate evaluation engine
- **Epic:** E-0007. **ADR:** 0028 (composes 0014, 0029).
- **Dep:** T-0601, T-0605, T-0610, T-0613.
- **Acceptance:** `Evaluate(ctx) Decision` with no side effects; composition
  order = reversibility floor → first-match policy → validation (for landing);
  precedence = irreversible/critical > `require_human` > `auto_approve` >
  `allow_claim` > deny; fired rule id + risk class + reversibility recorded.
- **Gate:** precedence unit tests — irreversible forces `critical`;
  `require_human` beats `auto_approve`; default-deny when no rule matches.

---

## Wave 2 — Run records, scheduler, actor-aware claim

### T-0420 (new) — Run records + lease-on-claim + actor snapshot
- **Epic:** E-0005 / E-0006 (carve). **ADR:** 0027.
- **Dep:** T-0410, T-0613.
- **Acceptance:** creating a run records `mode`, `actor_id`, `actor_snapshot_json`
  (captured at start), runtime/model, workspace, timestamps; claim + run + lease
  created in one transaction; snapshot immutable after registry edits.
- **Gate:** `-race` single-winner run creation; snapshot persisted/stable.

### T-0423 (new) — `runtime.Interface` + records-only stub
- **Epic:** E-0006 (carve). **ADR:** 0027.
- **Dep:** T-0420.
- **Acceptance:** minimal `internal/runtime.Interface`; stub emits synthetic
  lifecycle events (claimed→working→produced→awaiting-gate) and writes no files;
  selectable so Codex adapter (Phase 4) implements the same interface.
- **Gate:** stub unit test asserts events emitted, no filesystem writes.

### T-0422 (new) — `run_events` persistence + event hub
- **Epic:** E-0005. **ADR:** 0025, 0026.
- **Dep:** T-0420.
- **Acceptance:** run events persist to `run_events`; in-process pub/sub hub
  fans events to subscribers; slow subscriber dropped, never blocks producers.
- **Gate:** `-race` concurrent subscribe/publish; persistence round-trip.

### T-0402 — Scheduler loop (+ T-0404 dependency eligibility, folded)
- **Epic:** E-0005 (`T-0402`, `T-0404`). **ADR:** 0026, 0028.
- **Dep:** T-0420, T-0423, T-0614.
- **Acceptance:** serial tick finds eligible nodes (`ListEligible`, already
  dependency-aware), selects an actor, claims, and starts a supervisor; honors
  `MaxConcurrency`; lease heartbeat renews on interval against TTL; eligibility
  recomputes as dependencies complete.
- **Gate:** `-race` concurrency cap honored; integration test: dependency gates
  dispatch until prerequisite `done`.

### T-0606 — Actor-aware claim filtering
- **Epic:** E-0007 (`T-0606`). **ADR:** 0028, 0029.
- **Dep:** T-0402, T-0614.
- **Acceptance:** scheduler claim path runs actor selection through the gate
  engine; `work_type`/capability/policy mismatch denies the claim;
  `requested_actor` honored only as a hint and still policy-checked.
- **Gate:** mismatch-denies and hint-still-checked tests.

### T-0403 — Run pause/resume/cancel
- **Epic:** E-0005 (`T-0403`). **ADR:** 0027.
- **Dep:** T-0420.
- **Acceptance:** CLI/API request pause/resume/cancel; transitions audited;
  cancel releases the lease; `interrupted` is recovery-only, not client-settable.
- **Gate:** state-machine tests; cancel releases lease.

### T-0421 (new) — Checkpoint records + landing-squash semantics
- **Epic:** E-0006 (carve). **ADR:** 0027, 0015.
- **Dep:** T-0420.
- **Acceptance:** checkpoint *records* modeled; landing marks WIP checkpoints
  squashed and never on `main`; actual git commits + resume documented as Phase 4.
- **Gate:** record/model tests; landing sets squashed flag.

---

## Wave 3 — Approvals, decompose, escalation

### T-0602 — Approval records + decisions
- **Epic:** E-0007 (`T-0602`). **ADR:** 0028.
- **Dep:** T-0410, T-0614.
- **Acceptance:** approvals requested/listed/approved/rejected/clarified via CLI
  + API; records requesting actor, required actors/roles, deciding actor, reason;
  every decision audited; `require_human` cannot be bypassed.
- **Gate:** lifecycle tests; bypass-attempt rejected.

### T-0603 — Documentation auto-approval rule
- **Epic:** E-0007 (`T-0603`). **ADR:** 0028.
- **Dep:** T-0614, T-0602.
- **Acceptance:** docs-only internal-guidance change classifies low risk and
  auto-approves per policy; `land_to_main` stays human-gated in v1.
- **Gate:** docs change auto-approves; landing still requires human.

### T-0604 — Decompose gate + autonomy levels + SOP loading
- **Epic:** E-0007 (`T-0604`). **ADR:** 0011, 0028.
- **Dep:** T-0614, T-0602.
- **Acceptance:** `decompose` is a gated action, human-required in v1; work-type
  SOPs/context load from `.groundwork/sops/`; autonomy levels parsed but only a
  human can elevate.
- **Gate:** decompose requires human; SOP load + missing-SOP tests.

### T-0430 (new) — Decomposition proposal flow
- **Epic:** E-0005 / E-0007. **ADR:** 0030 (parent contract), 0028.
- **Dep:** T-0420, T-0602, T-0604.
- **Acceptance:** `POST /tickets/:id/decompose` opens a planning run; children
  created in `backlog`; proposal is an approval `type=decompose`; accept writes
  the parent `## Design / Contract` to canon and moves children to `todo` as
  dependencies allow; reject leaves children in `backlog`.
- **Gate:** end-to-end accept→children todo + parent contract written; reject path.

### T-0431 (new) — Escalation routing + human-gated re-plan
- **Epic:** E-0007 / E-0010. **ADR:** 0028 (resolves the [ADR 0024](../adr/0024-dependency-satisfaction-and-rollup-terminality.md) seam).
- **Dep:** T-0602.
- **Acceptance:** `POST /tickets/:id/escalate` records a typed escalation and
  moves the node to `blocked`; re-plan decision routed as approval `type=replan`;
  approved re-plan may send affected siblings to `rework` and unblock/re-point a
  dependent stalled behind a `cancelled` prerequisite.
- **Gate:** escalate blocks; replan recorded; stalled dependent resolvable.

---

## Wave 4 — Validation & canon

### T-0701 — Validation templates load
- **Epic:** E-0008 (`T-0701`). **ADR:** 0028. *(May merge into T-0601.)*
- **Dep:** T-0601.
- **Acceptance:** file patterns map to required checks (docs/Go/web examples in
  `policies.md`); `landing_risk_floor` honored.
- **Gate:** template-match unit tests.

### T-0702 — Record validation results
- **Epic:** E-0008 (`T-0702`).
- **Dep:** T-0410, T-0701.
- **Acceptance:** results stored and linked to ticket/run; artifact paths
  preserved; `GET /validation` + `gw validation list` expose them.
- **Gate:** store/API round-trip tests.

### T-0703 — Enforce landing validation gate
- **Epic:** E-0008 (`T-0703`). **ADR:** 0028.
- **Dep:** T-0614, T-0702.
- **Acceptance:** landing fails without required passing validation unless a human
  records an explicit, audited override.
- **Gate:** landing blocked without validation; override path audited.

### T-0440 (new) — Canon distillation + parent reconciliation
- **Epic:** E-0004 (fulfills `T-0306`). **ADR:** 0030, 0013, 0012.
- **Dep:** T-0430, T-0703.
- **Acceptance:** journal path/append API defined (tier-1, ignored); ratification
  hooks fire on decompose-accept / land / policy-SOP-approve; canon writes
  serialized through the coordinator (typed promotion, replace-in-place); parent
  reconciliation step at the composite; nothing written without ratification.
- **Gate:** ratification writes canon serially; non-ratified decision writes
  nothing; reconciliation invoked at parent.

---

## Wave 5 — SSE & recovery

### T-0803 — SSE event stream
- **Epic:** E-0009 (`T-0803`). **ADR:** 0025.
- **Dep:** T-0422.
- **Acceptance:** `GET /api/v1/events` streams hub events as SSE with heartbeat;
  `Last-Event-ID` resume tolerates reconnect.
- **Gate:** `httptest` receives events across a simulated reconnect.

### T-0901 — Startup reconciliation
- **Epic:** E-0010 (`T-0901`). **ADR:** 0027 (`recovery.md`).
- **Dep:** T-0420, T-0403.
- **Acceptance:** on boot, stale `running` runs without a live worker → marked
  `interrupted`; expired/orphaned leases released; dependency graph re-checked
  acyclic; worktree paths verified inside `.groundwork/worktrees/`.
- **Gate:** seeded stale runs/leases reconciled on startup.

### T-0902 (new) — Import ticket exports when SQLite is missing
- **Epic:** E-0010 (`T-0902`).
- **Dep:** existing `internal/exporter`.
- **Acceptance:** `gw ticket import` / cold start rebuilds node rows + dependency
  edges from committed Markdown exports; runtime leases are **not** restored.
- **Gate:** export → delete `state.sqlite` → import → identical node tree + edges.

---

## Build-first & end-to-end verification

1. **First:** T-0410 → T-0401 gives a running `gw server` with health +
   `/api/v1/state`; verify by extending `scripts/smoke.sh` to start/health/stop.
2. **Vertical slice (drives the whole spine via the records-only stub):**
   create ticket via API → scheduler claims an eligible leaf (single-winner,
   `-race`) → run record + lease → stub emits events → persisted + observed on
   SSE → node lands through the gate: `POST /land` opens a `land_to_main` approval
   via the gate engine (`require_human` in M2; docs auto-approval activates once
   the Phase 4 runtime supplies the changed-file `Scope`) → approving it enforces
   the validation gate and lands. `scripts/smoke.sh` exercises this end-to-end.
3. **Recovery:** kill + restart the server → stale run marked `interrupted`,
   lease expired (T-0901). Cold start: delete `state.sqlite`, restart → import
   rebuilds the tree (T-0902).

Wire this as a `scripts/smoke.sh` server section plus one coordinator integration
test using the stub runtime, so the coordinator is regression-covered before the
Codex runtime arrives in Phase 4.
