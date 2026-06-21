# Agent Quickstart

You are working in the Groundwork repository. **Phases 1 (CLI & Store) and 2
(Coordinator) are implemented and committed.** Phase 3 (self-hosting) is next.

## Read First

1. `README.md`
2. `AGENTS.md` (current boundary and what is out of bounds)
3. `docs/reference/architecture-map.md` (package layout)
4. `docs/product/roadmap.md` (phase status)
5. the live plan via `gw ticket tree` (the work tree is the planning source of truth, ADR 0040); see `.groundwork/WORKFLOW.md` for the loop
6. The specific contract or architecture doc for your task, and its governing ADR.

## What Exists

- `cmd/gw` + `internal/{cli,config,scaffold,store/sqlite,ticket,exporter,encoding,
  contextbrief,actor}` — the M1 CLI and pure-Go SQLite store.
- `internal/{server,client,scheduler,eventbus,run,runtime,policy,risk,approval,sop,
  canon}` — the M2 coordinator: `gw server` (HTTP API + SSE), scheduler, run records,
  gate engine, approvals, decompose/escalate, validation + landing gate, canon journal,
  recovery, import.

## Build & Verify

- `gofmt -l` clean; `CGO_ENABLED=0 go vet ./...`.
- `CGO_ENABLED=0 go build ./... && CGO_ENABLED=0 go test ./...`.
- `CGO_ENABLED=1 go test -race ./...` for concurrency paths.
- `bash scripts/smoke.sh` for the end-to-end binary-level flow.

## Out Of Bounds (see AGENTS.md)

- Codex runtime / real worktrees / real checkpoints (Phase 4) — M2 uses a records-only
  runtime stub.
- Autonomy elevation (Phase 5) — gates stay human-required in v1.
- Generated frontend assets — `docs/design/` is reference only; M2 is API/SSE only.
- Multi-human roles / auth / remote mode (post-v1).

## Essential Decisions

- Name: Groundwork. CLI: `gw`. Dot directory: `.groundwork/`. Language: Go.
- Store: SQLite operational state, ignored by default; durable committed state is docs,
  workflow, policies, ticket exports, and code.
- v1 server: localhost-only, single-user. v1 runtime target: Codex first (Phase 4).
- v1 landing and decomposition: human approval required, modeled as policy gates that
  loosen via SOPs/context/validation (Phase 5).
- Work hierarchy: uniform node tree; `kind` advisory, structure is leaf vs composite by
  triage; dependency-edge DAG overlay. A leaf is one verifiable change.

## Working Instruction

Keep changes small and verifiable. Build against the committed contracts; if a design
decision is missing or a contract must change, add or update an ADR before implementing.
Record new major decisions as ADRs (continue from 0031) and keep the reference docs
current so future sessions stay oriented.
