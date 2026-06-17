# Groundwork

Groundwork is a planned local-first orchestration system for managing coding agents against a single software project.

This repository is currently a **documentation-only bootstrap**. It intentionally contains no Go source code, no `go.mod`, no generated app assets, no `.groundwork/` runtime directory, and no SQLite database. The purpose of this first scaffold is to preserve the product, architecture, contracts, and planning decisions before implementation begins.

## What Groundwork Will Be

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

1. [AGENTS.md](AGENTS.md) for future agent operating instructions.
2. [docs/reference/agent-quickstart.md](docs/reference/agent-quickstart.md) for the compact implementation briefing.
3. [docs/product/vision.md](docs/product/vision.md) for product intent.
4. [docs/architecture/overview.md](docs/architecture/overview.md) for the architecture.
5. [docs/plan/work-tree.yaml](docs/plan/work-tree.yaml) for the initial implementation breakdown.

Implementation should begin in a fresh session from these docs. Do not infer missing design from this chat; record refinements in docs and ADRs.

