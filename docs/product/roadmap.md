# Roadmap

## Phase 0: Documentation Bootstrap

Create committed documentation that records product intent, architecture, contracts, ADRs, and initial work decomposition. Do not implement code.

**Status: complete.** The durable design — including the uniform work-node model, dynamic decomposition, dependency edges, escalation, SOPs/autonomy (ADRs 0009–0011), actors/work types (ADR 0023), and the web-surface design reference under `docs/design/` — is recorded. Implementation begins in a fresh session.

## Phase 1: CLI And Store

**Status: complete.** `gw init`, configuration loading, SQLite migrations, work-node CRUD (uniform nodes with advisory `kind` and leaf/composite triage), status transitions, dependency edges (DAG with cycle rejection), work-tree records and rollups, bounded context briefs (`gw context`), deterministic exports, transactional claim/lease with dependency-aware eligibility, and `gw status`/`board`/`doctor` are implemented. Decisions are recorded in ADRs 0016–0022.

## Phase 2: Coordinator

**Status: complete.** `gw server` (localhost HTTP API + SSE), the dependency- and
actor-aware scheduler over the transactional claim, run records (planning and
implementation modes) with actor snapshots and checkpoint records, decomposition
proposals, escalation/re-plan routing, approval records with actor and
reversibility gating composed through the gate engine, validation templates +
results + the landing gate, canon journal + ratification hooks, startup
reconciliation, and cold-start import are implemented. The Codex runtime is
modeled by a records-only stub; real agent execution is Phase 4. Decisions are
recorded in ADRs 0025–0031.

## Phase 3: Self-Host Low-Risk Work

Import the bootstrap work tree into Groundwork. Use Groundwork for docs, policies, and low-risk CLI/store tickets. Keep human landing approval.

## Phase 4: Codex Runtime

Add the Codex runtime adapter, isolated worktrees, run event streaming, transcripts, pause/resume, and tactical approval requests.

## Phase 5: Autonomy Path

Add stronger validation templates, risk scoring, policy learning suggestions, and progressively autonomous `execute`, `decompose`, review, and landing for explicitly permitted low-risk work as work-type SOPs and context mature.

## Phase 2 Product Features Beyond V1

- Chat approval adapter.
- Reviewer-agent approval.
- Policy learning installation flow.
- Earned and revocable autonomy from per-work-type outcome tracking (track record plus a circuit-breaker that demotes a class after bad outcomes).
- Budget gates (per-ticket and per-day token/cost ceilings that pause runs).
- Optional GitHub and Linear bridges.
- Optional remote or LAN mode with authentication.
