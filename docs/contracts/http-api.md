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
POST /api/v1/tickets/:id/decompose
POST /api/v1/tickets/:id/escalate
GET  /api/v1/tickets/:id/dependencies
POST /api/v1/tickets/:id/dependencies
DELETE /api/v1/tickets/:id/dependencies/:depId
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
