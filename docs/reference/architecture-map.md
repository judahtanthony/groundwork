# Architecture Map

```text
gw CLI
  -> coordinator/client service
  -> SQLite store

gw server
  -> scheduler (dependency-aware eligibility, triage dispatch)
  -> actor registry (local humans and AI actors, capabilities, runtime config)
  -> run supervisor (planning and implementation runs, checkpoints)
  -> context assembler (gw context briefs)
  -> durable decision/input/gate records + live approval/input projections
  -> approval router (actor + risk + reversibility gates)
  -> validation engine
  -> canon distiller (ratification-time promotion, parent reconciliation)
  -> exporters
  -> HTTP/SSE dashboard

agent runtime
  -> Codex adapter first
  -> isolated worktree
  -> event sink
  -> durable blocker/handoff records
  -> approval/input requests
```

## Core Directories

Packages built in Phase 1 (M1):

```text
cmd/gw                 # CLI entry point
internal/cli           # stdlib-flag command tree + router (ADR 0016)
internal/config        # root discovery + config.yaml schema (ADR 0021)
internal/scaffold      # `gw init` .groundwork tree
internal/store/sqlite  # connection, migrations, CRUD, deps, leases, eligibility
internal/ticket        # work-node domain model, status map, rollups
internal/exporter      # deterministic Markdown export (ADR 0020)
internal/encoding      # canonical timestamp/JSON encoding (ADR 0020)
internal/actor         # actor registry parsing + validation (ADR 0023; Phase 1.5)
internal/contextbrief  # bounded `gw context` brief assembly (ADR 0013)
```

Packages added in Phase 2 (M2):

```text
internal/server        # gw server HTTP API + SSE (ADR 0025)
internal/client        # coordinator HTTP client; store-vs-server boundary (ADR 0031)
internal/policy        # policy loading + gate evaluation engine (ADR 0028)
internal/risk          # risk score/class + reversibility classifier (ADR 0014)
internal/run           # run domain: status/mode state machines (ADR 0027)
internal/runtime       # runtime.Runtime seam + records-only stub (ADR 0027); codex adapter is Phase 4
internal/eventbus      # in-process pub/sub hub for SSE (ADR 0026)
internal/scheduler     # scheduler loop + actor selection + run supervisors (ADR 0026)
internal/approval      # approval domain; service lives in internal/server (ADR 0028)
internal/sop           # work-type SOP loading from .groundwork/sops/ (ADR 0011)
internal/canon         # journal + ratification hooks + parent reconciliation (ADR 0013/0030)
```

Ticket-attached decision records (ADR 0051) are the durable source for blockers,
input requests, rebuildable approval gates, rework requests, and recovery-needed
states; live approval/input queues are coordinator projections over those records.
Consequential decisions that need routing or canon output are normal work nodes
with decision-oriented `work_type`s (ADR 0052).

Validation has no package of its own: template matching lives in `internal/policy`
(`RequiredChecks`/`LandingRiskFloor`) and result records + the landing gate live in
`internal/store/sqlite` (`RecordValidation`, `Land`).

Packages added in Phase 3 (M3):

```text
internal/git           # minimal git-landing: stage + commit on the current branch (ADR 0034)
```

M3 added no other package: bootstrap import reuses `internal/exporter` + the cold-start
importer (`internal/cli`), the human path uses existing transitions/gates with AI claims
gated by trust policy, context-miss capture lives in `internal/canon` (`Miss`/`Misses`),
and land-time staging lives in `internal/cli` (`resolveLandStaging`). The commit itself
is made by the coordinator (`internal/server`) after a landing gate completes.

Package areas expected in later phases (not yet created):

```text
internal/worktree
internal/runtime/codex
internal/ui
```

Phase 1 folds the planned `internal/context` (the `gw context` brief) into
`internal/cli`; it can be extracted if the assembler grows. Keep new package
boundaries aligned with this map.
