# Groundwork

Groundwork is a local-first orchestration system for managing coding agents against a single software project.

## Status

Phase 1 (CLI & Store) and Phase 2 (Coordinator) are implemented and committed: the `gw`
CLI, the pure-Go SQLite store, `gw server` (localhost HTTP API + SSE), the dependency-
and actor-aware scheduler, run records, the trust/risk/reversibility gate engine,
approvals, decomposition and escalation/re-plan flows, the validation + landing gate,
canon journal + ratification hooks, and recovery/import. The Codex runtime is a
records-only stub pending Phase 4. Next: Phase 3 (self-hosting). See
[docs/product/roadmap.md](docs/product/roadmap.md) and [docs/plan/milestones.md](docs/plan/milestones.md).

## What Groundwork Is

- Project name: Groundwork.
- CLI name: `gw`.
- Managed-project dot directory: `.groundwork/`.
- Initial implementation language: Go.
- Runtime target for v1: Codex first, with a runtime interface kept open for future adapters.
- Storage model: SQLite is the local operational store during runtime.
- Durability model: committed docs, workflow, policies, ticket exports, and code are durable project state; SQLite, worktrees, run transcripts, raw logs, generated views, and approval inbox projections are ignored by default.
- Server model: localhost-only and single-user in v1.
- UI model: Go server-rendered HTML with minimal JavaScript in v1; optional TypeScript frontend later only if needed.
- Work model: a uniform tree of nodes (kind is advisory; structure is leaf vs composite, decided by triage at claim time), with a dependency-edge DAG overlay.
- Decomposition: composite nodes decompose just-in-time into children as a reviewable proposal; revisions propagate upward via escalation.
- Autonomy model: landing and decomposition are human-gated in v1 but modeled as policy gates that loosen as SOPs, context, and validation mature.

## Why It Exists

Groundwork aims to make agent-managed software work transparent, local, low-cost, and low-lock-in. It borrows the broad operating idea from OpenAI Symphony: humans should manage work, constraints, validation, trust, and visibility while agents increasingly execute tasks end-to-end.

The v1 trust boundary is conservative. Human approval is required before landing code to `main`. That approval is modeled as a policy gate so the system can later support autonomous landing for low-risk, well-validated work.

## How To Read This Repo

Start here:

1. [AGENTS.md](AGENTS.md) for agent operating instructions and the current boundary.
2. [docs/reference/agent-quickstart.md](docs/reference/agent-quickstart.md) for the compact briefing and build/verify commands.
3. [docs/product/vision.md](docs/product/vision.md) for product intent.
4. [docs/architecture/overview.md](docs/architecture/overview.md) for the architecture.
5. [docs/reference/architecture-map.md](docs/reference/architecture-map.md) for the package layout.
6. [docs/plan/work-tree.yaml](docs/plan/work-tree.yaml) for the implementation breakdown.

Work against the committed docs and ADRs; do not infer missing design from chat history. Record refinements as ADRs and keep the reference docs current.

