# ADR 0033: Human-Performed Execution via Manual Status Transitions in M3

Status: Accepted

## Context

M3 makes Groundwork manage its own low-risk docs/CLI work, but agent **execution is
human-performed** — the Codex runtime is Phase 4. The M2 coordinator already has the
scheduler, transactional claim, leases, run records, and a `runtime.Interface` whose
only M3 implementation is the records-only `Stub`, which emits synthetic lifecycle
events and writes nothing ([ADR 0027](0027-run-lifecycle-and-checkpoint-records.md)).
[ADR 0026](0026-coordinator-concurrency-model.md)'s scheduler → run → events → gate →
landing loop is already exercised end-to-end against that stub in M2.

The question: should a human's work in M3 flow through a coordinator run (claim → run
→ a completion signal) or through manual status transitions?

## Decision

**M3 humans drive work via manual status transitions.** A human works however they
like (by hand or with a coding tool of their choice), then advances the node with
`gw ticket transition` (`todo` → `in_progress` → `review`) and `gw ticket land`. The
run/dispatch machinery is **not** used for human work in M3.

Rationale: no executor exists in M3, so routing a human through a run would simulate
an executor that isn't there while requiring net-new surface — a blocking/manual
`runtime.Runtime`, a run-completion CLI/API command, and a human `allow_claim`
policy rule — to obtain coverage the M2 stub tests already provide. The gates that
matter for M3 are **action-triggered, not run-triggered**: `gw ticket land` opens
the `land_to_main` approval and enforces the validation gate
([ADR 0028](0028-gate-evaluation-engine.md)) regardless of whether a run exists.
`human.owner` already holds `approve`/`review: ["*"]` in the scaffolded
`actors.yaml`, so the human landing-approval path needs no new policy.

Because the M2 scheduler auto-claims eligible `todo` nodes and dispatches them to
AI actors through the runtime stub, running it would race the human for any node.
`gw server --no-scheduler` therefore runs the coordinator — API, gates, approvals,
landing — without the scheduler, so the human owns the lifecycle. The scheduler
remains on by default; `--no-scheduler` is the M3 human-work mode.

## Consequences

M3 adds no runtime, scheduler, or CLI surface for the human execution path, and the
Phase 3/4 boundary stays crisp. Run-based dispatch gets its real workout in Phase 4,
where a human "build session" finally drives a *real* executor — exactly what
T-1003 ("first Codex-assisted ticket") is. Status transitions are still recorded in
`audit_events` for traceability. The accepted trade-off is that human work in M3
carries no run record (no token/time telemetry); this is fine because the work is
human and out-of-band. Should run records on human work ever be wanted, they compose
with the same gate engine without revisiting this decision.
