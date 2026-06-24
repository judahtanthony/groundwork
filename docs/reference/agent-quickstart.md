# Agent Quickstart

You are working in the Groundwork repository. **Phases 1 (CLI & Store), 2
(Coordinator), and 3 (self-host low-risk work) are implemented and committed.**
Phase 4 is next: the operator UI for ticket visibility, ready/blocked work,
approval decisions, and landing preview. Phase 5 follows with bounded autonomy
and bulk review for manual and agent-directed work; durable async handoff and the
real Codex runtime move to Phase 6.

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
  recovery, import. ADRs 0051/0052 now specify the Phase 6 durable async handoff model:
  ticket sidecar decision records are authoritative for rebuildable blockers, and
  consequential decisions are normal routed work nodes.
- `internal/git` plus the committed `.groundwork/tickets/` exports — M3
  self-hosting: Groundwork manages its own work tree, low-risk work is
  human-performed via manual transitions, and landing is a real coordinator-made
  git commit.

## Build & Verify

- `gofmt -l` clean; `CGO_ENABLED=0 go vet ./...`.
- `CGO_ENABLED=0 go build ./... && CGO_ENABLED=0 go test ./...`.
- `CGO_ENABLED=1 go test -race ./...` for concurrency paths.
- `bash scripts/smoke.sh` for the end-to-end binary-level flow.

## Out Of Bounds (see AGENTS.md)

- Durable async handoff and real Codex worktree execution are Phase 6; Phase 4
  should not depend on them.
- Broad autonomy elevation beyond approved envelopes — Phase 5 may reduce manual
  approval overhead through envelopes, reviewer checks, and bulk review, but
  never self-elevate.
- Full SPA polish and future admin-agent surfaces — Phase 4 may build the
  minimum operator UI needed for tickets, approvals, and land preview, but rich
  admin/chat surfaces remain out of scope.
- Multi-human roles / auth / remote mode (post-v1).

## Essential Decisions

- Name: Groundwork. CLI: `gw`. Dot directory: `.groundwork/`. Language: Go.
- Store: SQLite operational state, ignored by default; durable committed state is docs,
  workflow, policies, ticket exports, ticket sidecar decision records, and code.
- Durable state is filesystem-authoritative; SQLite is a rebuildable projection plus
  ephemeral runtime coordination store (ADR 0053).
- v1 server: localhost-only, single-user. v1 runtime target: Codex first in Phase 6.
- v1 landing and decomposition: human approval required by default, modeled as
  policy gates that loosen only through explicit revised Phase 5 envelope/policy
  work.
- Work hierarchy: uniform node tree; `kind` advisory, structure is leaf vs composite by
  triage; dependency-edge DAG overlay. A leaf is one verifiable change.

## Working Instruction

Keep changes small and verifiable. Build against the committed contracts; if a design
decision is missing or a contract must change, add or update an ADR before implementing.
Record new major decisions as ADRs (continue from 0031) and keep the reference docs
current so future sessions stay oriented.
