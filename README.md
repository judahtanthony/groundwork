# Groundwork

Groundwork is a local-first orchestration system for managing coding agents against a single software project.

## Status

Phase 1 (CLI & Store) and Phase 2 (Coordinator) are implemented and committed: the `gw`
CLI, the pure-Go SQLite store, `gw server` (localhost HTTP API + SSE), the dependency-
and actor-aware scheduler, run records, the trust/risk/reversibility gate engine,
approvals, decomposition and escalation/re-plan flows, the validation + landing gate,
canon journal + ratification hooks, and recovery/import. Phases 1–3 are complete.
Phase 4 is the operator UI, Phase 5 is bounded autonomy and bulk review, and
Phase 6 replaces the records-only runtime stub with durable handoff and the real
Codex runtime. See [docs/product/roadmap.md](docs/product/roadmap.md).

## What Groundwork Is

- Project name: Groundwork.
- CLI name: `gw`.
- Managed-project dot directory: `.groundwork/`.
- Initial implementation language: Go.
- Runtime target for v1: Codex first, with a runtime interface kept open for future adapters.
- Storage model: SQLite is the local operational store during runtime.
- Durability model: committed docs, workflow, policies, ticket exports, and code are durable project state; SQLite, worktrees, run transcripts, raw logs, generated views, and approval inbox projections are ignored by default.
- Server model: localhost-only and single-user in v1.
- UI model: localhost operator web UI in v1; start with the minimum ticket/approval
  surfaces needed to unblock the human, then grow toward the embedded static SPA
  described in ADR 0042.
- Work model: a uniform tree of nodes (kind is advisory; structure is leaf vs composite, decided by triage at claim time), with a dependency-edge DAG overlay.
- Decomposition: composite nodes decompose just-in-time into children as a reviewable proposal; revisions propagate upward via escalation.
- Autonomy model: landing and decomposition are human-gated by default; Phase 5
  adds bounded approval envelopes and bulk review before Phase 6 background
  runtime execution.

## Why It Exists

Groundwork aims to make agent-managed software work transparent, local, low-cost, and low-lock-in. It borrows the broad operating idea from OpenAI Symphony: humans should manage work, constraints, validation, trust, and visibility while agents increasingly execute tasks end-to-end.

The v1 trust boundary is conservative. Human approval is required before landing code to `main`. That approval is modeled as a policy gate so the system can later support autonomous landing for low-risk, well-validated work.

## Build And Run

Groundwork is a single Go binary with no cgo dependency for normal use. The `Makefile`
wraps the canonical commands:

```sh
make build            # CGO_ENABLED=0 go build -o bin/gw ./cmd/gw  ->  ./bin/gw
make install          # CGO_ENABLED=0 go install ./cmd/gw          ->  gw on $PATH
```

Then drive a project with the `gw` CLI:

```sh
gw init               # scaffold .groundwork/ in the current repo
gw ticket import      # rebuild the runtime store from committed exports (cold start)
gw ticket tree        # inspect the work tree (the planning source of truth)
gw status             # node counts by status, plus root rollups
gw server             # run the localhost coordinator: HTTP API + SSE, approvals, runs, landing
```

`state.sqlite` is runtime-only and git-ignored; on a fresh checkout `gw ticket import`
rebuilds it byte-for-byte from the committed Markdown exports under `.groundwork/tickets/`
([ADR 0020](docs/adr/0020-canonical-encoding-deterministic-export.md)). The CLI is self-documenting — run
`gw help`, or `gw <command> -h` for any command.

## Development

Every change must pass the same gates `scripts/smoke.sh` and CI enforce. Each has a
`Makefile` target:

```sh
make fmt     # gofmt -l .                       (must print nothing)
make vet     # CGO_ENABLED=0 go vet ./...
make test    # CGO_ENABLED=0 go test ./...
make race    # CGO_ENABLED=1 go test -race ./... (the race detector requires cgo)
make smoke   # bash scripts/smoke.sh            (end-to-end CLI smoke test)
```

Groundwork manages its own development: the work tree is the planning source of truth
([ADR 0040](docs/adr/0040-groundwork-is-planning-source-of-truth.md)) and low-risk work is
taken through Groundwork itself, with a human landing gate. See the
[self-hosting runbook](docs/reference/self-hosting.md) and
[.groundwork/WORKFLOW.md](.groundwork/WORKFLOW.md) for the operating loop.

## How To Read This Repo

Start here:

1. [AGENTS.md](AGENTS.md) for agent operating instructions and the current boundary.
2. [docs/reference/agent-quickstart.md](docs/reference/agent-quickstart.md) for the compact briefing and build/verify commands.
3. [docs/product/vision.md](docs/product/vision.md) for product intent.
4. [docs/architecture/overview.md](docs/architecture/overview.md) for the architecture.
5. [docs/reference/architecture-map.md](docs/reference/architecture-map.md) for the package layout.
6. The live plan in Groundwork: run `gw ticket tree` (the work tree is the planning source of truth, [ADR 0040](docs/adr/0040-groundwork-is-planning-source-of-truth.md)); see [.groundwork/WORKFLOW.md](.groundwork/WORKFLOW.md).

Work against the committed docs and ADRs; do not infer missing design from chat history. Record refinements as ADRs and keep the reference docs current.
