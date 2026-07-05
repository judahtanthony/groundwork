# ADR 0051: Async Agent Handoff And Durable Decision Records

Status: Accepted
Implemented: Partial

## Context

Groundwork is moving from human-performed ticket transitions toward autonomous
background agents. Those agents should generally run without waiting on a live
human. If they are blocked, the correct behavior is to make the blocker durable,
release the lease, and let the scheduler spend capacity on other eligible work.
A later run may be performed by a different actor, model, or runtime instance.

The current model over-weights live runtime rows for some blockers. Pending
approvals live in SQLite, and proposal payloads for `decompose` or `replan` can
be lost if `.groundwork/state.sqlite` is purged and rebuilt from ticket exports.
That can strand a ticket in `blocked` or `review` with no visible pending gate.
`land_to_main` is often recoverable by rerunning the landing command, but
`decompose` and `replan` can lose semantic payload that cannot be reconstructed
from status alone.

## Decision

Model sessions are disposable. Raw transcripts, tool events, and run logs are
supporting evidence, not authoritative memory. The authoritative durable memory
for active work is split as follows:

- The ticket/work tree is the durable operational source of truth for work,
  status, blockers, dependency edges, handoff context, and ticket-attached
  decision/request history.
- Canon remains the durable knowledge source: ADRs, architecture docs,
  contracts, policies, SOPs, and product docs.
- Runs remain attempt history: events, transcript, tool calls, artifacts, diffs,
  checkpoint refs, token/cost metadata, and evidence supporting audit.

Anything that explains why a ticket is `blocked`, in `review`, or ready for
`rework` must be exported with the ticket before the worker exits.

Ticket directories gain durable, structured sidecar records for ticket-attached
requests and decisions. The accepted v1 sidecar is:

```text
.groundwork/tickets/<ticket-id>/ticket.md
.groundwork/tickets/<ticket-id>/decisions.ndjson
```

Each line in `decisions.ndjson` is a deterministic JSON object in canonical key
order. The file records append-only semantic events such as:

- `decision_requested`
- `decision_responded`
- `input_requested`
- `input_answered`
- `approval_requested`
- `approval_decided`
- `rework_requested`
- `recovery_needed`

The durable request/decision id is stable across rebuilds. Runtime approval ids
are handles for the live coordinator queue and may change after rebuild.

Approvals are not removed. They are reframed:

- A durable ticket decision/gate record is the source of truth when a pending
  gate must survive rebuild.
- The `approvals` table and approval endpoints are live coordinator projections
  over pending durable records.
- A purely tactical, safely replayable runtime wait may remain runtime-only, but
  any gate whose loss would strand or alter work must have a durable record.

Startup/cold rebuild must:

1. Import ticket exports and ticket sidecar decision records.
2. Rebuild the work tree and dependency edges.
3. Recreate pending approval, input, and decision queues from durable pending
   records.
4. Detect inconsistent states, including blocked tickets with no durable blocker,
   review tickets with no recoverable proposal/gate record, and runtime-only
   approvals whose payload was required but lost.
5. Create or surface an explicit `recovery_needed` record for unrecoverable lost
   context instead of silently stranding the ticket.

Recovery cannot reconstruct context that was never made durable. It may rerun
planning/action, move the ticket to `rework`, create a recovery ticket, ask a
human to decide manually, or recreate an approval from durable state when safe.

## Async Blocked-Run Lifecycle

When an autonomous agent is blocked, it should:

1. Checkpoint worktree state when applicable.
2. Write run events.
3. Write transcript and artifacts as runtime evidence.
4. Write a durable ticket-attached request/decision record if the block must
   survive rebuild.
5. Optionally create a dependent decision ticket when the question is
   consequential (ADR 0052).
6. Move the original ticket to `blocked` with an explainable blocker.
7. Release the lease and end the run.
8. Let the scheduler dispatch other eligible work.

When the response arrives:

1. The response is recorded on the durable ticket-attached record.
2. Live queues update.
3. Any dependent decision ticket may complete, updating canon or the parent
   contract.
4. The original ticket becomes eligible when dependencies and durable blockers
   are satisfied.
5. A new run starts from a structured resume packet assembled from durable state.

Resume usually means "start a new run/turn from durable context," not "continue
the exact old model session." A resume packet should include the current ticket
context, ancestor contract, acceptance criteria, dependency status, relevant
resolved decision/input records, latest rework notes, prior run handoff summary,
current diff/checkpoint refs, validation state, artifacts/links, relevant
transcript excerpts only, and an explicit next recommended action.

## Consequences

The sidecar format is specified in `docs/contracts/decision-records.md` and
referenced by `ticket-export.md`, `file-layout.md`, `sqlite-schema.md`,
`http-api.md`, and `cli.md`. `state-model.md`, `runtime-model.md`,
`agent-runtime.md`, `recovery.md`, and `trust-and-approvals.md` are updated so
SQLite is a live index/projection for rebuildable queues rather than the source
of truth for durable blockers.

This ADR refines ADR 0012 without changing its master test: if losing a pending
request would strand or alter the work, it is not ephemeral. It also refines ADR
0013: ticket-attached decision records are durable operational memory, while
canon remains the durable knowledge memory after ratification.
