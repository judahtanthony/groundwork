# Milestones

## M0: Documentation Bootstrap — complete

All durable product, architecture, contract, ADR, reference, and planning docs exist, including the uniform work-node model, dynamic decomposition, dependency edges, escalation, SOPs/autonomy (ADRs 0009–0011), and the web-surface design reference under `docs/design/`. No code exists.

## M1: CLI And Store Foundation — complete

`gw init`, config loading, SQLite migrations, work-node CRUD, triage, status transitions, dependency edges (DAG with cycle rejection), work-tree rollups, bounded `gw context` briefs, deterministic Markdown export, transactional claim/lease with dependency-aware eligibility, and `gw doctor` exist. Phase 1 decisions are recorded in ADRs 0016–0022. Implementation lives under `cmd/gw` and `internal/`; see `docs/reference/architecture-map.md`.

## M2: Coordinator Foundation — complete

`gw server`, dependency- and actor-aware claims, leases, run records (planning and implementation), actor snapshots, decomposition proposals, escalation/re-plan, approvals with reversibility + actor gating, validation records and the landing gate, canon journal + ratification hooks, startup reconciliation, cold-start import, and SSE/API basics exist. The Codex runtime is a records-only stub (Phase 4 replaces it). Decisions are recorded in ADRs 0025–0031.

## M3: Self-Hosting Preparation

The bootstrap work tree is imported into Groundwork. Groundwork manages low-risk docs and CLI tasks.

## M4: Codex Runtime

Codex runs execute in isolated worktrees, stream events, request approvals, and produce transcripts.

## M5: Autonomy Path

Risk scoring, validation confidence, actor-aware policy suggestions, and low-risk autonomous `execute`, `decompose`, review, and landing are available under explicit policy as work-type SOPs mature.
