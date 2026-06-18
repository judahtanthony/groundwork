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
```

Package areas expected in later phases (not yet created):

```text
internal/git
internal/scheduler
internal/actor
internal/worktree
internal/runtime/codex
internal/approval
internal/policy
internal/sop
internal/risk
internal/canon
internal/checkpoint
internal/validation
internal/server
internal/ui
```

Phase 1 folds the planned `internal/context` (the `gw context` brief) into
`internal/cli`; it can be extracted if the assembler grows. Keep new package
boundaries aligned with this map.
