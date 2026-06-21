# Agent Instructions For Groundwork

**Phases 1–3 are complete and committed** — M1 (CLI & store), M2 (coordinator), and M3
(self-host low-risk work). The committed docs and ADRs remain the source of truth for
design; code is written against them. **Phase 4 (Codex Runtime, M4)** is next.

## Required Reading

Before proposing or making implementation changes, read:

1. `docs/reference/agent-quickstart.md`
2. `docs/reference/architecture-map.md`
3. `docs/reference/conventions.md`
4. `docs/plan/work-tree.yaml`

For deeper context, read the matching architecture, contract, or ADR file before changing that area.

## Current Boundary

Phase 4 (M4, Codex runtime) is next: replace the records-only run stub with the real
Codex adapter — agent execution in isolated worktrees, run-event streaming and
transcripts, and real git checkpoints/squash/resume — then run the first Codex-assisted
ticket (T-1003) through Groundwork. Build against the committed contracts and
architecture docs; where a contract must change, record an ADR rather than diverging
silently.

What already exists (build on it, do not rebuild):

- **M1** — the `gw` CLI and pure-Go SQLite store.
- **M2** — the coordinator: `gw server` (localhost HTTP API + SSE), the dependency- and
  actor-aware scheduler over the transactional claim, run records + lifecycle, the gate
  engine (trust + reversibility + risk + actor policy), approvals, decomposition and
  escalation/re-plan flows, validation records + the landing gate, the canon journal +
  ratification hooks, startup reconciliation, and cold-start import.
- **M3 (self-hosting)** — Groundwork manages its own work tree: the bootstrap tree is
  imported as committed Markdown ticket exports (ADR 0032); low-risk docs work runs
  through Groundwork human-performed via manual status transitions, with AI claims gated
  by the trust policy (`allow_claim`) rather than auto-dispatched (ADR 0033); landing is a
  real git commit the coordinator makes on the current branch via the minimal
  `internal/git` (ADR 0034); and context-misses feed the canon/brief loop (ADR 0035).

Still out of bounds until their phase begins:

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
distillation, server-vs-store boundary). Phase 3 decisions are recorded in ADRs
0032–0035 (bootstrap import via authored Markdown, human execution via manual
transitions, minimal git-landing, context-miss capture).

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

Groundwork now manages its own work tree: the bootstrap `docs/plan/work-tree.yaml` has been imported as managed ticket exports under `.groundwork/tickets/` (the YAML is the historical bootstrap, no longer the live plan). Inspect and evolve the plan through `gw` (`gw ticket tree`, `gw ticket create`, …). Keep tickets small enough to validate independently.

Create ADRs for any new major decisions and update the condensed reference docs so future sessions can quickly understand the system.

