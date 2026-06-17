# Agent Instructions For Groundwork

This repository has entered **Phase 1 (CLI & Store)** implementation. The Phase 0
documentation bootstrap is complete and remains the source of truth for design; code
is now being written against it. The documentation-only boundary has been lifted.

## Required Reading

Before proposing or making implementation changes, read:

1. `docs/reference/agent-quickstart.md`
2. `docs/reference/architecture-map.md`
3. `docs/reference/conventions.md`
4. `docs/plan/work-tree.yaml`

For deeper context, read the matching architecture, contract, or ADR file before changing that area.

## Current Boundary

Phase 1 implementation is active. Creating `go.mod`, `cmd/`, `internal/`, SQLite
databases, and migrations is now expected work. Build against the committed contracts
and architecture docs; where a contract must change, record an ADR rather than diverging
silently. Phase 1 scope is the CLI and store (steps 1–7 of
`docs/reference/implementation-guide.md`): module + `gw` skeleton, config discovery,
`gw init`, SQLite migrations + store, node CRUD + triage + deterministic export,
work-tree records + dependency edges + rollups + `gw context`, and transactional
claim/lease. The coordinator/server, runs, dashboard, escalation routing, canon
distillation, reversibility gating, and checkpoints remain **Phase 2** and should not be
built yet.

Still out of bounds until their phase begins:

- `gw server`, runs, scheduler, approvals, validation gates (Phase 2).
- Generated frontend assets — `docs/design/` holds the web-surface visual reference
  (the Claude Design wireframe handoff) and the decomposition UI spec. It is reference
  only; do not turn the prototypes into generated frontend assets until web-surface work
  is explicitly started.

Phase 1 decisions are recorded in ADRs 0016–0022 (CLI framework, SQLite driver,
migrations, ticket IDs, encoding/export determinism, config discovery, status model).

## Design Commitments To Preserve

- Project name: Groundwork.
- CLI: `gw`.
- Managed-project state directory: `.groundwork/`.
- Language: Go for v1.
- Runtime target: Codex first, adapter interface preserved.
- SQLite is the operational store during runtime.
- SQLite is ignored by default.
- Durable committed state: docs, workflow, policies, ticket exports, and code.
- Runtime ignored state: SQLite, WAL/SHM, worktrees, run transcripts, raw command logs, generated views.
- Server is localhost-only and single-user in v1.
- Leaf nodes represent one verifiable change.
- Work is a uniform tree of nodes; kind is advisory, structure is leaf vs composite decided by triage at claim time.
- Composite nodes decompose just-in-time into children; decomposition is a reviewable proposal.
- Dependency edges form a DAG overlay; nodes are eligible only when dependencies are satisfied.
- Revisions propagate upward via escalation; re-plan is human-gated in v1.
- Human approval is required in v1 for both landing and decomposition, but both are policy gates that loosen via SOPs/context/validation to permit future autonomy.

## Planning Source

Until Groundwork can manage itself, use `docs/plan/work-tree.yaml` as the initial breakdown. Keep tickets small enough to validate independently.

When implementation starts, create ADRs for any new major decisions and update the condensed reference docs so future sessions can quickly understand the system.

