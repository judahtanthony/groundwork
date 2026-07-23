# ADR 0060: Review Rejection Routes To `todo` With A Rework Decision Record (Retire The `rework` Status)

Status: Proposed
Implemented: Not implemented

Amends: [ADR 0022](0022-phase1-status-model.md) (status set and transition map),
`docs/architecture/state-model.md`, and `docs/contracts/decision-records.md`
(rebuild semantics).

## Context

When a run completes, the scheduler moves its leaf to `review`. A reviewer then
accepts the work (`review → approved → landing → done`) or sends it back. The status
model inherited from `docs/architecture/state-model.md` and
[ADR 0022](0022-phase1-status-model.md) includes a `rework` status for the "send it
back" case, with edges `review → rework` and `rework → {in_progress, review, cancelled}`.
ADR 0022 explicitly deferred the authorization of the rework cycle to "Phase 2".

Dogfooding surfaced that the rework cycle is a **dead end** as wired today:

1. **No re-dispatch path.** Dispatch requires `todo`: `txClaim`
   (`internal/store/sqlite/lease.go`) returns `ErrNotEligible` for any non-`todo`
   ticket, and `IsEligible`/`ListEligible` (`internal/store/sqlite/eligibility.go`)
   filter to `todo`. The transition map has **no `rework → todo` edge**, so a reworked
   ticket can never re-enter the dispatchable set — the scheduler, `gw run once`, and
   `gw ticket claim` all refuse it. Returning a rejected leaf to work today requires
   walking `review → rework → in_progress → blocked → todo` by hand.
2. **No feedback capture.** `gw ticket transition` takes only `<id> <status>`; there is
   nowhere to record *why* the work was rejected or *what to change*. The prompt builder
   (`internal/resume/resume.go`) recognizes `status == "rework"` but can only emit the
   generic string `"address rework feedback and re-submit for review"` — it reads no
   feedback, so the re-dispatched agent is told it is doing rework without being told
   what to fix.

The core realization: **a status label cannot carry the correction.** The unique context
a re-dispatched implementation agent needs (for example, "build in `web/`, not the
server-rendered templates") is *feedback content*. That content must be captured as a
durable record regardless of which status the ticket sits in. Once that record exists, a
second dispatchable status (`rework`) is redundant on top of it — and expensive, because
it would force "dispatchable = {`todo`, `rework`}" through eligibility, `txClaim`,
rollups, `gw next`, and every UI surface.

**The channel already exists and was never wired.** `decisions.ndjson` is the durable
per-ticket record of gated request→authority→resolution events
(`docs/contracts/decision-records.md`). It already declares `rework_requested` as a
required event type; `internal/decision/decision.go` already defines
`EventReworkRequested`; the contract's rebuild semantics already describe projecting
those records. **No production code emits or consumes one.** The contract's own opening
states decision records explain why a ticket is "blocked, in review, or ready for
rework". A new parallel sidecar would duplicate a channel purpose-built for this, and
would split one decision point across two timelines — since the *accept* path already
produces a landing approval on `decisions.ndjson`.

## Decision

**Retire the explicit `rework` status. A review rejection routes the leaf
`review → todo` and writes a `rework_requested` decision record carrying the correction,
which the resume packet surfaces in the next run's prompt.**

1. **Transition.** Replace the `review → rework` and `rework → *` edges with a single
   guarded `review → todo` edge. Remove `rework` from the status set. `todo` remains the
   *only* dispatchable status, so eligibility, `txClaim`, the scheduler, and rollups are
   unchanged.

2. **Reuse the decision channel — no new sidecar.** The rejection is recorded on the
   existing `.groundwork/tickets/<id>/decisions.ndjson` as a `rework_requested` event,
   using the existing schema: `statement` (the correction, required and human-readable),
   `handoff_summary`, `ticket_id`, `run_id`, `decided_by`, `decided_at`, and optionally
   `related_files` / `follow_up`. No new file, encoding, or reader is introduced.

3. **Dedicated verb, mandatory note, atomic write.** Expose rejection as a reviewer-role
   verb (for example `gw review reject <id> --note …`), not a bare `transition`. The note
   is required, and the `rework_requested` record is written **in the same transaction**
   as `review → todo` — so the scheduler, which claims a `todo` leaf as soon as it is
   eligible, can never re-dispatch before the correction is durably present. This race is
   real: an eligible leaf is picked up on the next tick.

4. **Prompt wiring.** `resume.Assemble` already ingests decision records. It reads the
   latest unaddressed `rework_requested` for the ticket and sets `NextAction` /
   `HandoffSummary` to that record's `statement`, so the agent's prompt states the
   concrete correction instead of a generic rework string. A `rework_requested` record is
   considered addressed once a later run on that ticket reaches `review`.

5. **Provenance without a status.** The transition emits a `review.rejected` audit event.
   "Was this reworked, and how many times" is derived from the `rework_requested` records
   and audit events. A UI may render an "In rework" affordance *computed* from `todo` +
   an unaddressed `rework_requested` record, and/or a `rework_count`; it is not a status.

6. **Update the coder SOPs.** Implementers of this ADR **must** update the SOPs an
   implementing actor follows — `.groundwork/sops/technical_implementation/SOP.md` and
   `.groundwork/sops/test_implementation/SOP.md` — so a re-dispatched agent handles
   returned work correctly. The SOPs must direct the actor to:
   - **Check for an unaddressed `rework_requested` record when orienting.** The existing
     "Orient before touching code" step already reads prior decisions via
     `gw ticket context <id>`; make the rework case explicit rather than incidental.
   - **Treat a returned node as a continuation, not a fresh start.** Address the specific
     correction in `statement`; do not restart from scratch, re-litigate the approach, or
     discard prior work that was not objected to.
   - **Honor a correction that narrows scope or surface.** When the note constrains
     *where* the change belongs, that constraint binds as tightly as the envelope's file
     scope; widening it is a boundary crossing to escalate, not to work around.
   - **State how the feedback was addressed** in the completion summary, so the next
     reviewer can verify the correction directly.

   The `documentation` SOP gets the same orientation guidance if it materially differs.

7. **Update the contract.** `docs/contracts/decision-records.md` rebuild semantics
   currently read "`rework_requested` records explain why a ticket is in `rework`".
   Since `rework` no longer exists, this becomes: they explain why a ticket was returned
   to `todo`, and an unaddressed record marks work awaiting correction.

If a future need arises for returned work to be gated *differently* from fresh `todo`
(for example, held from auto-dispatch pending explicit human re-release, or routed to a
different actor), encode it as a boolean flag on the `todo` ticket (`awaiting_rework`),
not a resurrected status.

## Consequences

- The status machine loses a state and its dead-end edges; `todo` stays the single
  dispatchable status, so dispatch, eligibility, and rollup surfaces are untouched.
- Rejection becomes usable end-to-end for the first time: reject-with-note re-dispatches
  the *same* ticket carrying the correction, replacing today's manual four-hop status
  walk and avoiding cancel-and-recreate.
- Review feedback becomes first-class durable data on the channel already built for it,
  available to prompts, SOPs, the review UI, audit, and cold rebuild — with one timeline
  per ticket covering both the accept and reject paths.
- `docs/architecture/state-model.md` and the ADR 0022 transition map must drop `rework`
  and add the guarded `review → todo` edge. Code referencing `rework` is removed or
  recomputed from data — notably the `status == "rework"` branch in
  `internal/resume/resume.go`.
- Because the correction reaches the agent through the SOPs and the resume packet, the
  SOP updates in decision 6 are part of the change, not a follow-up.
- Out of scope for the Web UI epic (T-1036); tracked as a platform ticket under T-1074.
