# ADR 0054: Approval Envelopes v1

Status: Pending Review
Implemented: Not started

Refines the exploration in [ADR 0044](0044-hierarchical-planning-and-approval-envelopes.md)
into a concrete, minimal v1 contract for Phase 5. Composes with
[ADR 0028](0028-gate-evaluation-engine.md) (gates), [ADR 0038](0038-authority-as-loosenable-gate.md)
(authority as policy), [ADR 0040](0040-groundwork-is-planning-source-of-truth.md) /
[ADR 0053](0053-filesystem-authoritative-durable-state.md) (file-authoritative state),
[ADR 0046](0046-resource-scope-ownership-and-conflict-policy.md) (scope), and
[ADR 0058](0058-integration-targets-and-landing-levels-v1.md) (integration target / git
landing mechanics).

## Context

Just-in-time decomposition is correct for uncertain work, but approving every child
turns the human into the serialization point (ADR 0043). Phase 5's job is to move human
review to **boundaries**: approve a bounded parent/root contract once, then let child
planning and execution proceed inside it until something exits the boundary.

This must deliver value **before the Phase 6 runtime exists**: even when children are
executed manually (a human directing Claude Code in the coding role), one approved
envelope replaces a stream of per-child tactical approvals.

## Decision

Add an **approval envelope** as a durable record attached to a composite or root node. A
human approves it once; it is then a necessary authorization input for AI claims, child
execution, and child landing within the subtree (the gate composition is
[ADR 0056](0056-envelope-aware-claim-authorization.md)).

### v1 Envelope Shape

```yaml
approval_envelope:
  id: env-T-2000
  node_id: T-2000
  status: active            # active | revoked | superseded
  approved_by: human.owner
  approved_at: 2026-06-25T00:00:00Z
  approved_actions:         # subset of the action vocabulary (ADR 0028/0038)
    - decompose_children
    - execute_children
    - land_children_to_parent
    - replan_within_goal
  planning:
    max_depth: 2
    max_children: 12
    allowed_work_types: [technical_design, technical_implementation, test_implementation, documentation]
  scope:
    files:
      allow:          [internal/runtime/**, internal/worktree/**]
      require_review: [docs/adr/**, .groundwork/policies/**]
      deny:           [.env*, "**/*secret*"]
  validation:
    required_templates: [go]
  risk_ceiling: medium      # risk class above which work exits the envelope
  allowed_roles: [planner, coding, reviewer]
  escalation:
    on_unexpected_files: true
    on_contract_change: true
    on_validation_failure: true
    on_risk_above_ceiling: true
    on_public_api_change: true
```

v1 deliberately supports **file-glob scope only**. Symbol- and named-resource scope
(ADR 0046) are deferred until the file path proves out.

### What an envelope authorizes / does not

It authorizes, *within its limits*: child decomposition to `max_depth`, execution of
child leaves whose work type and scope match, child landing to the **parent integration
target**, and bounded re-planning that stays inside the approved goal.

It never authorizes (these remain human-gated, ADR 0043/0038): landing the root to
`main`, policy changes, autonomy elevation, irreversible actions, touching `deny`/unknown
files, parent-contract changes, work above `risk_ceiling`, or proceeding on failed
validation. Any of these raises an **exception approval** rather than being silently
allowed or denied.

### Storage — file-authoritative

The envelope is durable operational state, so it is stored as a node-attached sidecar in
the ticket export — `.groundwork/tickets/<id>/envelope.yaml` — and mirrored into SQLite
for live evaluation (ADR 0040/0053). This resolves ADR 0044's open question toward
durability and dogfood transparency: an envelope is reviewable as a committed file and
survives a store rebuild.

### Lifecycle

1. Planner proposes a parent contract + envelope (a `decompose`/`approve_envelope`
   request through the existing approval service; ADR 0030).
2. Policy determines the required approver role(s) from work type, risk, scope, and action
   (ADR 0055).
3. A human (or, later, an authorized role) approves, rejects, or asks for clarification.
4. On approval, children are created in **`backlog`**; the planner marks each `todo` when
   ready (resolves ADR 0044's open question conservatively — approval grants a boundary,
   not an immediate dispatch wave).
5. A child action that exceeds the envelope opens an **exception approval** scoped under
   the parent envelope (queues group by parent).
6. Parent integration verifies child summaries, diffs, and validation before the final
   parent/root review ([ADR 0057](0057-bulk-review-bundles-v1.md)).

### Revocation

An envelope may be revoked. v1 (pre-runtime): revocation flips `status` to `revoked`,
which immediately blocks new claims/landings under it; there are no live background runs
to stop. When the Phase 6 runtime lands, revocation also signals active runs to checkpoint
and hand off (ADR 0047 blocked-run handoff). A superseding re-plan marks the old envelope
`superseded` and links the replacement.

## Consequences

- New durable record type (envelope) with a sidecar export and SQLite mirror; the approval
  service gains an envelope-approval path; the gate gains envelope inputs (ADR 0056).
- Pre-runtime value: one human approval authorizes a bounded batch of manual child work.
- The envelope is the unit a future runtime, reviewer agent, and bulk-review bundle all
  reference.

## Open Questions

- Should `max_children`/`max_depth` be hard caps or soft (warn + exception)? v1 proposes
  hard caps that raise an exception when exceeded.
- Should envelopes nest (a child composite holding its own envelope inside a parent's)? v1
  allows at most one active envelope per ancestor chain to keep evaluation simple; nesting
  is a later decision.
- How are `require_review` paths mapped to roles before multi-human identity exists? v1:
  they require `owner` review (ADR 0055).
