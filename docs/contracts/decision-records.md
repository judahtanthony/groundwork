# Ticket Decision Records Contract

Ticket-attached decision records are durable operational memory. They explain
why a ticket is blocked, in review, or ready for rework, and they let pending
input/approval/decision queues rebuild after `.groundwork/state.sqlite` is
purged.

## Location

```text
.groundwork/tickets/<ticket-id>/decisions.ndjson
```

The file is optional when a ticket has no durable request/decision history. When
present, it is committed with the ticket export.

## Encoding

The file is newline-delimited JSON. Each line is one semantic event. Export must
use the same canonical JSON encoding rules as ticket exports: stable key order,
UTC timestamps, no insignificant whitespace, and deterministic ordering by
`created_at`, then `id`, then `sequence`.

The durable `id` is stable across SQLite rebuilds. Runtime approval ids or queue
ids are live handles and may change after rebuild.

## Event Types

Required event types:

- `decision_requested`
- `decision_responded`
- `input_requested`
- `input_answered`
- `approval_requested`
- `approval_decided`
- `rework_requested`
- `recovery_needed`

## Common Fields

```json
{
  "id": "D-0007",
  "sequence": 7,
  "event_type": "approval_requested",
  "ticket_id": "T-1234",
  "run_id": "R-4567",
  "request_type": "decompose",
  "work_type": "technical_design",
  "status": "pending",
  "requested_by": "ai.codex.default",
  "requested_at": "2026-06-24T15:04:05Z",
  "requested_actor": "human.owner",
  "required_roles": ["owner"],
  "policy_inputs": {
    "action": "decompose",
    "risk_class": "medium",
    "reversible": true
  },
  "statement": "Accept the proposed child tickets and parent contract?",
  "handoff_summary": "The parent is too broad to implement safely as one leaf.",
  "options": [
    {
      "id": "accept",
      "label": "Accept proposal"
    },
    {
      "id": "revise",
      "label": "Request rework"
    }
  ],
  "recommendation": "accept",
  "risk_notes": "Documentation-only plan change; reversible by editing ticket exports.",
  "scope_notes": "Creates child tickets under T-1234.",
  "related_files": [".groundwork/tickets/T-1234/ticket.md"],
  "artifacts": [".groundwork/runs/R-4567/artifacts/proposal.json"],
  "checkpoints": [],
  "depends_on": [],
  "response": null,
  "decided_by": null,
  "decided_at": null,
  "follow_up": [],
  "canon_updates_needed": ["docs/architecture/runtime-model.md"]
}
```

Fields may be omitted only when not applicable. `ticket_id`, `event_type`,
`status`, actor/timestamp fields for the event, and a human-readable `statement`
or `handoff_summary` are required.

## Status

Valid statuses:

- `pending`
- `answered`
- `accepted`
- `rejected`
- `superseded`
- `recovered`

The current status for a durable request is the latest non-superseded event for
that `id`.

## Rebuild Semantics

Cold rebuild imports `ticket.md` and `decisions.ndjson` together. Pending durable
records recreate live coordinator queues:

- pending `approval_requested` records project into the approvals queue,
- pending `input_requested` records project into the input queue,
- pending `decision_requested` records either keep the ticket blocked or point to
  a dependent decision ticket,
- `rework_requested` records explain why a ticket is in `rework`,
- `recovery_needed` records surface lost runtime context that cannot be safely
  reconstructed.

If a ticket status implies a blocker or proposal but no durable record explains
it, startup reconciliation must create or surface `recovery_needed` rather than
silently leaving the ticket stranded.
