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
modeled by a records-only stub; real agent execution is Phase 6. Decisions are
recorded in ADRs 0025–0031.

## Phase 3: Self-Host Low-Risk Work

Import the bootstrap work tree into Groundwork. Use Groundwork for docs, policies, and low-risk CLI/store tickets. Keep human landing approval. Implemented; decisions are recorded in ADRs 0032–0035.

## Superseded Phase 4: Codex Runtime

This was the original post-M3 plan: add the Codex runtime adapter, isolated
worktrees, run event streaming, transcripts, pause/resume, and tactical approval
requests.

**Status: superseded before implementation.** The Codex runtime remains required,
but recent planning showed that the human approval bottleneck should be reduced
first. The runtime work moves to Phase 6, after the operator UI and bounded
autonomy / bulk-review model are in place.

## Superseded Phase 5: Autonomy Path

This was the original autonomy plan: add stronger validation templates, risk
scoring, policy learning suggestions, and progressively autonomous `execute`,
`decompose`, review, and landing for explicitly permitted low-risk work as
work-type SOPs and context mature.

**Status: superseded before implementation.** The direction remains valid, but it
is now framed as bounded parent/root delegation, role-specific agents, durable
handoffs, reviewer-agent checks, and bulk human review rather than generic
autonomy loosening.

## Phase 4: Operator UI

Make Groundwork useful as the human operator surface before adding more
background execution machinery.

- Build the minimum operator web UI needed now: ticket visibility, ready/blocked
  queues, approval inbox, approve/reject/clarify actions, and landing diff
  preview. This is the urgent slice of the broader web UI plan in ADR 0042; full
  SPA polish can follow later.
- Keep every mutation routed through the existing coordinator APIs and gates; the
  UI is an operator client, not a policy bypass.
- Defer durable async handoff, file-authoritative state, and the real Codex
  runtime to Phase 6.

The live execution plan is represented by the Phase 4 tickets in Groundwork,
especially `T-1060`, `T-1036`, and the operator-unblock slice under `T-1061`.

## Phase 5: Bounded Autonomy And Bulk Review

Reduce approval overhead by moving human review to approved parent/root
boundaries and exception points while preserving auditability, validation,
default-deny authorization, and human control of high-risk authority changes.
This phase deliberately comes before the real Codex runtime because the same
model helps even while work is manually directed through Claude Code: the human
can approve an envelope, review summarized evidence, and avoid approving every
tactical step.

- Add role-specific agents: planner, coding, and reviewer actors.
- Add approved envelopes for parent/root work: allowed actions, work types,
  actors, file/resource scopes, validation requirements, risk ceilings, and
  exception triggers.
- Add manual-mode bulk-review flow: child summaries, diffs, validation evidence,
  reviewer findings, unresolved exceptions, and a final parent/root review
  bundle.
- Loosen `allow_claim` only where the execution substrate exists; before Phase 6,
  envelopes and reviewer-agent checks reduce human overhead without implying
  unattended background execution.
- Keep root/main landing, policy changes, autonomy elevation, irreversible
  actions, failed validation, and unexpected scope expansion human-gated in v1.

The live execution plan is represented by the revised Phase 5 tickets in
Groundwork. The exploratory direction lives in draft ADRs 0043-0050; the
implementation-ready v1 contracts are ADRs 0054-0058 (approval envelopes,
role-aware actors, envelope-aware claim authorization, bulk review bundles, and
integration targets / landing levels).

## Phase 6: Durable Handoff And Codex Runtime

Add the runtime/backend execution substrate after the operator UI and bounded
review model have reduced manual approval overhead.

**Status: complete.** Durable async handoff (ADR 0051) and filesystem-authoritative
durable state (ADR 0053) are implemented: ticket-attached `decisions.ndjson`
sidecars, queue rebuild + recovery-needed detection, store-level write-through with
divergence detection, consequential-decision routing (ADR 0052), blocked-run
handoff/resume packets, completion-summary requirements + staleness, and durable
context in briefs. The records-only stub is replaced by the Codex adapter (ADR 0027)
running in isolated per-run worktrees (ADR 0059): config-selected launch, event
streaming to SQLite + `events.ndjson`, worktree diff capture feeding gates, WIP
checkpoints squashed into the integration branch at landing, and resume-from-
checkpoint. The Phase 5 seams are activated: the scheduler routes AI claims through
envelope-aware authorization (ADR 0056) and the envelope file-scope + escalation
triggers are enforced against the real diff. A capstone end-to-end test drives an
implementation ticket through the runtime with the human landing gate intact. A new
ADR 0059 records the worktree-per-run topology. Decisions are recorded in ADR 0059
and the existing ADRs 0051/0053/0052/0027/0015/0056. The live plan is `T-1071`, with
`T-1052` (durable handoff/state), `E-0006` (Codex runtime), and `E-0012`
(envelope/escalation activation).

## Phase 7+ Product Features Beyond V1

These remain future work after the revised v1 Phase 4-6 path.

- Chat approval adapter.
- Full embedded SPA polish beyond the operator-unblock slice.
- Reviewer-agent approval beyond bounded reviewer checks.
- Policy learning installation flow.
- Earned and revocable autonomy from per-work-type outcome tracking (track record plus a circuit-breaker that demotes a class after bad outcomes).
- Budget gates (per-ticket and per-day token/cost ceilings that pause runs).
- Optional GitHub and Linear bridges.
- Optional remote or LAN mode with authentication.
