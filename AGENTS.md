# Agent Instructions For Groundwork

**Phase 1 (CLI & Store, M1)** and **Phase 2 (Coordinator, M2)** are complete and
committed. The committed docs and ADRs remain the source of truth for design; code is
written against them. **Phase 3 (Self-Host Low-Risk Work, M3)** is next.

## Required Reading

Before proposing or making implementation changes, read:

1. `docs/reference/agent-quickstart.md`
2. `docs/reference/architecture-map.md`
3. `docs/reference/conventions.md`
4. `docs/plan/work-tree.yaml`

For deeper context, read the matching architecture, contract, or ADR file before changing that area.

## Current Boundary

Phase 3 (M3, self-hosting) is next: import the bootstrap work tree
(`docs/plan/work-tree.yaml`) into Groundwork and run low-risk docs and CLI/store
tickets *through* Groundwork itself, keeping human landing approval. Build against the
committed contracts and architecture docs; where a contract must change, record an ADR
rather than diverging silently.

What already exists (build on it, do not rebuild): the `gw` CLI and pure-Go SQLite store
(M1), and the M2 coordinator — `gw server` (localhost HTTP API + SSE), the dependency-
and actor-aware scheduler over the transactional claim, run records + lifecycle, the
gate engine (trust + reversibility + risk + actor policy), approvals, decomposition and
escalation/re-plan flows, validation records + the landing gate, the canon journal +
ratification hooks, startup reconciliation, and cold-start import.

Still out of bounds until their phase begins:

- The **Codex runtime** — real agent execution, isolated worktrees, run-event streaming,
  transcripts, and real git checkpoints/resume (Phase 4). M2 ships a records-only runtime
  stub; do not launch Codex yet.
- **Autonomy elevation** — earned/auto loosening of gated actions (Phase 5). Gates stay
  human-required in v1; never self-elevate.
- **Generated frontend assets** — `docs/design/` holds the web-surface visual reference
  (the Claude Design wireframe handoff) and the decomposition UI spec. It is reference
  only; the M2 dashboard surface is API/SSE only. Do not turn the prototypes into
  generated frontend assets until web-surface work is explicitly started.
- **Multi-human roles / authentication / remote mode** (post-v1).

Phase 1 decisions are recorded in ADRs 0016–0022 (CLI framework, SQLite driver,
migrations, ticket IDs, encoding/export determinism, config discovery, status model).
Phase 2 decisions are recorded in ADRs 0025–0031 (HTTP/SSE transport, coordinator
concurrency, run lifecycle/checkpoint records, gate engine, actor identity, canon
distillation, server-vs-store boundary).

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

