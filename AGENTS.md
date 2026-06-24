# Agent Instructions For Groundwork

**Phases 1–3 are complete and committed** — M1 (CLI & store), M2 (coordinator), and M3
(self-host low-risk work). The committed docs and ADRs remain the source of truth for
design; code is written against them. **Phase 4** is next: operator UI for
ticket/approval visibility, ready/blocked work, approval decisions, and landing
preview. Phase 5 is bounded autonomy and bulk review; Phase 6 is durable async
handoff + filesystem-authoritative ticket state and the real Codex runtime.

## Required Reading

Before proposing or making implementation changes, read:

1. `docs/reference/agent-quickstart.md`
2. `docs/reference/architecture-map.md`
3. `docs/reference/conventions.md`
4. `.groundwork/WORKFLOW.md`, then the live plan via `gw ticket tree`

For deeper context, read the matching architecture, contract, or ADR file before changing that area.

## Current Boundary

Phase 4 is next: make the operator web UI useful enough to see tickets, ready/blocked
work, approval requests, approval decisions, and landing previews. Phase 5 follows with
bounded autonomy and bulk review so approval envelopes and summarized evidence reduce
manual approval overhead even before unattended background execution. Phase 6 then
implements durable async handoff, filesystem-authoritative ticket state, and the real
Codex adapter — agent execution in isolated worktrees, run-event streaming and
transcripts, and real git checkpoints/squash/resume — then runs the first
Codex-assisted ticket (T-1003) through Groundwork. Build against the committed contracts
and architecture docs; where a contract must change, record an ADR rather than diverging
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

- **Broad autonomy elevation beyond approved envelopes** — Phase 5 may reduce manual
  approval overhead through envelopes, reviewer checks, and bulk review. Gates stay
  human-required until explicitly loosened through policy and approved envelopes; never
  self-elevate.
- **Durable handoff and real Codex execution** — Phase 6 work. Phase 4 operator UI and
  Phase 5 bulk-review work should not require the real runtime to exist.
- **Full SPA polish and future admin-agent surfaces** — `docs/design/` holds the
  web-surface visual reference (the Claude Design wireframe handoff) and the
  decomposition UI spec. Phase 4 may build the minimum operator UI needed for
  tickets, approvals, and landing preview; richer SPA polish follows the web-surface
  tickets and ADR 0042.
- **Multi-human roles / authentication / remote mode** (post-v1).

Phase 1 decisions are recorded in ADRs 0016–0022 (CLI framework, SQLite driver,
migrations, ticket IDs, encoding/export determinism, config discovery, status model).
Phase 2 decisions are recorded in ADRs 0025–0031 (HTTP/SSE transport, coordinator
concurrency, run lifecycle/checkpoint records, gate engine, actor identity, canon
distillation, server-vs-store boundary). Phase 3 decisions are recorded in ADRs
0032–0035 (bootstrap import via authored Markdown, human execution via manual
transitions, minimal git-landing, context-miss capture). ADRs 0036–0038 record the
long-term direction: work as a universal substrate (0036), transitional defaults vs.
architectural invariants (0037), and authority as a uniform loosenable gate (0038).

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
- Leaf nodes represent one verifiable change.
- Work is a uniform tree of nodes; kind is advisory, structure is leaf vs composite decided by triage at claim time.
- Composite nodes decompose just-in-time into children; decomposition is a reviewable proposal.
- Dependency edges form a DAG overlay; nodes are eligible only when dependencies are satisfied.
- Revisions propagate upward via escalation.
- Auditability, default-deny authorization, reversibility-as-a-tracked-input, canon-as-memory, deterministic export, and the single serialized coordinator are architectural invariants (ADR 0037): autonomy may remove the human but never these.

### Transitional Defaults (loosenable, not commitments)

These are conservative *current settings* that loosen as SOPs, validation, context, and
trust mature — not permanent guarantees (see [ADR 0037](docs/adr/0037-transitional-defaults-vs-invariants.md)
and [ADR 0038](docs/adr/0038-authority-as-loosenable-gate.md)):

- Server is localhost-only and single-user in v1.
- Human approval is required for landing and decomposition; both are policy gates.
- Re-plan / escalation acceptance is human-gated.
- Irreversible actions and autonomy elevation are human-gated; under ADR 0038 both are loosenable policy defaults, not structural floors.

## Planning Source

Groundwork manages its own work tree: it is the planning source of truth (ADR 0040). The bootstrap `work-tree.yaml` and the static phase-by-phase breakdowns have been imported and retired (see git history); the live plan is the managed ticket exports under `.groundwork/tickets/`. Inspect and evolve it through `gw` (`gw ticket tree`, `gw ticket create`, `gw ticket context`, …) per `.groundwork/WORKFLOW.md`. Keep tickets small enough to validate independently.

Create ADRs for any new major decisions and update the condensed reference docs so future sessions can quickly understand the system.
