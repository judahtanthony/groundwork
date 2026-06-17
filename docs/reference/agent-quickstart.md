# Agent Quickstart

You are working in the Groundwork repository. This repo is currently a documentation-only bootstrap.

## Read First

1. `README.md`
2. `AGENTS.md`
3. `docs/reference/architecture-map.md`
4. `docs/plan/work-tree.yaml`
5. The specific contract or architecture doc for your task.

## Do Not Do Unless Explicitly Asked

- Do not create `go.mod`.
- Do not create Go source code.
- Do not create `cmd/` or `internal/`.
- Do not create `.groundwork/`.
- Do not create SQLite files.
- Do not implement runnable logic.

## Essential Decisions

- Name: Groundwork.
- CLI: `gw`.
- Dot directory: `.groundwork/`.
- Language: Go.
- Store: SQLite operational state, ignored by default.
- Durable committed state: docs, workflow, policies, ticket exports, code.
- Runtime ignored state: SQLite, WAL/SHM, worktrees, run transcripts, raw logs, generated views.
- v1 server: localhost-only and single-user.
- v1 runtime: Codex first.
- v1 landing: human approval required, modeled as policy gate.
- v1 decomposition: human approval required, modeled as policy gate that loosens via SOPs/context/validation.
- Work hierarchy: uniform node tree; kind advisory, structure is leaf vs composite by triage; dependency-edge DAG overlay.
- Leaf node: one verifiable change.

## Fresh Implementation Session Instruction

Start with one ticket from `docs/plan/work-tree.yaml`. Keep the change small and verifiable. If a design decision is missing, add or update an ADR before implementing.

