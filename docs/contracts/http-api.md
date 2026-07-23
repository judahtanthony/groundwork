# HTTP API Contract

The v1 server binds to `127.0.0.1:4500` by default and is single-user.

## Required Endpoints

```text
GET  /api/v1/state
GET  /api/v1/readiness
GET  /api/v1/tickets
POST /api/v1/tickets
GET  /api/v1/tickets/:id
PATCH /api/v1/tickets/:id
POST /api/v1/tickets/:id/claim
POST /api/v1/tickets/:id/transition
POST /api/v1/tickets/:id/triage
POST /api/v1/tickets/:id/reparent
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
GET  /api/v1/tickets/:id/land/route
POST /api/v1/tickets/:id/land
POST /api/v1/tickets/:id/land-to-parent
POST /api/v1/tickets/:id/envelope
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
GET  /api/v1/settings
POST /api/v1/settings/agents-md/sync
POST /api/v1/doctor
GET  /api/v1/events
```

`POST /api/v1/tickets/:id/decompose` opens a planning run; the resulting decomposition proposal is decided through the approvals endpoints (`approve` accepts the proposal, `clarify` asks the agent for more detail). `approve`/`reject` cover the `decompose`, landing, and tactical gates uniformly.

`GET /api/v1/readiness` returns the operator's current next/ready/blocked view:
`{"next":{"ticket": …, "brief": …}|null, "ready":[…], "blocked":[…]}`.
`ready` is the eligible set (`todo` with dependencies satisfied) in the same
ancestor-priority value order used by `gw next`, `gw ticket list --ready`, and
the scheduler. `next` is its first node plus the bounded context brief returned
by the ticket context endpoint. Each `blocked` entry is a `todo` node excluded
by unmet dependencies and includes `blocked_by: [{"id", "status"}]`.

`POST /api/v1/tickets/:id/claim` mirrors the guided human claim: it verifies the
node is `todo` with all dependencies satisfied, assigns the requested actor (default
`human.owner`), and transitions it to `in_progress`. `POST
/api/v1/tickets/:id/triage` classifies the node as `leaf` or `composite`, and
`POST /api/v1/tickets/:id/reparent` moves it under the supplied `parent` (an empty
parent makes it a root); both retain the store's cycle and validation checks.

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
`POST /api/v1/tickets/:id/land-to-parent` lands a child onto its root integration
branch (ADR 0058): it marks the child done and commits its export plus staged work
to that branch — distinct from `land_to_main`, which is the human-gated root merge.
It enforces the same validation gate as `land_to_main` (a child with a failing
validation result is refused).
`GET /api/v1/tickets/:id/land/route` reports how `gw ticket land` should land the
node — `{"route": "parent"|"main", "integration_branch", "run_branch"}`. A node
whose nearest integration target is an ancestor lands to that branch (`parent`);
a root that owns its integration branch, or an unparented node, lands to main
(`main`). The CLI reads this so a plain `gw ticket land <child>` auto-routes a
run-backed child to `land_to_parent` — squashing its `gw/run/<id>` branch — instead
of committing the main working tree and orphaning the run's code (ADR 0058).
`POST /api/v1/tickets/:id/envelope` proposes an approval envelope for the node
(ADR 0054): the body is the draft envelope JSON (`approved_actions`,
`allowed_roles`, `planning`, `scope`, `risk_ceiling`, …) and it opens a pending
human-gated `approve_envelope` approval. Approving that approval activates the
envelope (authoritative sidecar + mirror) and establishes the root integration
branch. An empty or unrecognized `approved_actions` set returns `400 invalid_actions`.
This is the production entry point for `gw envelope propose`.
`POST /api/v1/tickets/:id/land` drives landing through the `land_to_main` approval
gate (ADR 0028): policy auto-approves and lands immediately, otherwise it returns
`{"landed": false, "approval": …}` for a human to approve (approving lands).
`{"override": true}` lands immediately, bypassing both the approval and validation
gates with an audited override. These were added in M2; the changed-file set that
selects template-required checks and enables docs auto-approval of landing is
supplied by the Phase 6 runtime.

### Policies

`GET /api/v1/policies` reads the committed policy files and returns the parsed
trust and validation policies. `rules` is the UI-oriented ordered trust-rule
view; each item is `{"group", "order", "rule"}`, where `group` is
`require_human`, `auto_approve`, or `allow_claim`, `order` is one-based within
that group, and `rule.id` is the durable stable id. `validation_templates` is
sorted by template name and contains `{"name", "template"}` entries with file
globs, required checks, and any landing risk floor. Parse/load warnings are
returned in `warnings`.

`PUT /api/v1/policies` requests replacement of the complete ordered trust
policy. The body is:

```json
{
  "ticket_id": "T-1045",
  "trust": {
    "schema": "groundwork_trust_policy/v1",
    "require_human": [],
    "auto_approve": [],
    "allow_claim": []
  }
}
```

The referenced ticket makes the change auditable. The replacement must parse
successfully and preserve every existing stable rule id, group, and order. A
valid request returns `202` with `{"approval": …}` for an `amend_policy` gate;
`trust.yaml` is not changed until that approval is accepted. Acceptance writes
the file atomically and reloads the coordinator's trust-policy view.

`GET /api/v1/policies/suggestions` lists pending policy-learning suggestions;
`?status=all` includes prior decisions. `POST .../:id/promote` and `POST
.../:id/dismiss` record the human review outcome and return the updated
suggestion. As with `gw policy promote`, promotion records intent only and never
self-elevates or rewrites autonomy policy.

Actor endpoints expose the current local actor registry from `.groundwork/actors.yaml`. Runs expose actor ids and snapshots through the run endpoints; snapshots are runtime audit records, not edits to the actor registry.

### Settings and diagnostics

`GET /api/v1/settings` returns the resolved repository, SQLite, and config paths;
the configured server address split into bind and port; the agent engine, optional
model, and sandbox mode; maximum concurrency and lease TTL/heartbeat; and AGENTS.md
sync status (`missing`, `out_of_sync`, or `synced`).

`POST /api/v1/settings/agents-md/sync` creates or refreshes a marker-delimited
Groundwork section in the repository's `AGENTS.md`. Text outside that managed
section is preserved. The response is the updated sync status.

`POST /api/v1/doctor` runs the same project, configuration, actor-registry, and
database checks as `gw doctor --json` and returns
`{"healthy": bool, "checks": [{"name", "status", "detail"}]}`. Diagnostic
failures are represented by `healthy: false` and error-status checks rather than
an HTTP error.

`GET /api/v1/runs/:id` returns the run record plus run-detail evidence:
`plan` (plan/triage/decomposition events), `changed_files`, `validations`, an optional
linked `approval`, and optional `cost`. Token metrics remain the run record's
`input_tokens`, `output_tokens`, and `total_tokens` fields. `GET
/api/v1/runs/:id/events` reads `.groundwork/runs/<run-id>/events.ndjson` through
the coordinator and returns events oldest-first as
`{"id","run_id","event_type","message","payload","created_at"}`. A missing log
returns an empty array. Malformed JSON records, a torn trailing append, and
individual oversized records are skipped without failing the response; genuine
filesystem/read errors return `500 transcript_read_failed`.

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
