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
  -> approval router (actor + risk + reversibility gates)
  -> validation engine
  -> canon distiller (ratification-time promotion, parent reconciliation)
  -> exporters
  -> HTTP/SSE dashboard

agent runtime
  -> Codex adapter first
  -> isolated worktree
  -> event sink
  -> approval requests
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

Validation has no package of its own: template matching lives in `internal/policy`
(`RequiredChecks`/`LandingRiskFloor`) and result records + the landing gate live in
`internal/store/sqlite` (`RecordValidation`, `Land`).

Package areas expected in later phases (not yet created):

```text
internal/git
internal/worktree
internal/runtime/codex
internal/ui
```

Phase 1 folds the planned `internal/context` (the `gw context` brief) into
`internal/cli`; it can be extracted if the assembler grows. Keep new package
boundaries aligned with this map.
