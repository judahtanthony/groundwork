# HTTP API Contract

The v1 server binds to `127.0.0.1:4500` by default and is single-user.

## Required Endpoints

```text
GET  /api/v1/state
GET  /api/v1/tickets
POST /api/v1/tickets
GET  /api/v1/tickets/:id
PATCH /api/v1/tickets/:id
POST /api/v1/tickets/:id/transition
GET  /api/v1/tickets/:id/children
GET  /api/v1/tickets/:id/context
GET  /api/v1/tickets/:id/decisions
POST /api/v1/tickets/:id/decisions
POST /api/v1/tickets/:id/decompose
POST /api/v1/tickets/:id/escalate
GET  /api/v1/tickets/:id/dependencies
POST /api/v1/tickets/:id/dependencies
DELETE /api/v1/tickets/:id/dependencies/:depId
GET  /api/v1/tickets/:id/validations
POST /api/v1/tickets/:id/validations
GET  /api/v1/tickets/:id/land/preview
POST /api/v1/tickets/:id/land
GET  /api/v1/runs
POST /api/v1/runs
GET  /api/v1/runs/:id
GET  /api/v1/runs/:id/events
POST /api/v1/runs/:id/pause
POST /api/v1/runs/:id/resume
POST /api/v1/runs/:id/cancel
GET  /api/v1/actors
GET  /api/v1/actors/:id
POST /api/v1/actors/validate
GET  /api/v1/approvals
GET  /api/v1/approvals/:id
POST /api/v1/approvals/:id/approve
POST /api/v1/approvals/:id/reject
POST /api/v1/approvals/:id/clarify
GET  /api/v1/policies
PUT  /api/v1/policies
GET  /api/v1/policies/suggestions
POST /api/v1/policies/suggestions/:id/promote
POST /api/v1/policies/suggestions/:id/dismiss
GET  /api/v1/events
```

`POST /api/v1/tickets/:id/decompose` opens a planning run; the resulting decomposition proposal is decided through the approvals endpoints (`approve` accepts the proposal, `clarify` asks the agent for more detail). `approve`/`reject` cover the `decompose`, landing, and tactical gates uniformly.

`GET/POST /api/v1/tickets/:id/decisions` reads and appends durable ticket-attached
decision records (`decision-records.md`). Pending durable `approval_requested` and
`input_requested` records are projected into live coordinator queues; approval ids may
change after rebuild, but the durable request id remains stable.

`GET /api/v1/tickets/:id/validations` lists recorded validation results and
`POST /api/v1/tickets/:id/validations` records one (the coordinator-mediated path
`gw validation run` uses, so the server's state/SSE stay coherent — ADR 0031).
`GET /api/v1/tickets/:id/land/preview` returns the staged change set a landing of
the node would commit — `{"id", "staged", "diff"}` — the server-mediated read of
`gw ticket land --preview` (ADR 0034/0041). It is read-only (no staging, commit,
or approval) and returns `400 not_a_repo` outside a git work tree.
`POST /api/v1/tickets/:id/land` drives landing through the `land_to_main` approval
gate (ADR 0028): policy auto-approves and lands immediately, otherwise it returns
`{"landed": false, "approval": …}` for a human to approve (approving lands).
`{"override": true}` lands immediately, bypassing both the approval and validation
gates with an audited override. These were added in M2; the changed-file set that
selects template-required checks and enables docs auto-approval of landing is
supplied by the Phase 6 runtime.

Actor endpoints expose the current local actor registry from `.groundwork/actors.yaml`. Runs expose actor ids and snapshots through the run endpoints; snapshots are runtime audit records, not edits to the actor registry.

`GET /api/v1/events` should use Server-Sent Events for realtime dashboard updates.

## Error Shape

Use a consistent JSON error envelope:

```json
{
  "error": {
    "code": "not_found",
    "message": "Ticket not found"
  }
}
```
