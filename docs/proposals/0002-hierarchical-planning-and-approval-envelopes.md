# Proposal 0002: Hierarchical Planning And Approval Envelopes

Status: Draft

## Goal

Reduce planning and landing approval load without weakening quality by allowing humans to
approve a bounded parent or root contract. Children may be planned and executed inside
that approved envelope until they hit an escalation trigger.

## Problem

Groundwork's just-in-time decomposition is right for uncertain work, but approving every
small child can turn the human into the serialization point. Planning an entire tree up
front can also create false precision. Complex software work often requires discovery,
design, implementation, testing, and integration feedback before the next useful layer is
known.

## Proposal

Add the concept of an **approval envelope** to a root or composite node. The envelope is
a reviewed boundary within which planning and execution can proceed.

An envelope can authorize:

- decomposition to a specified depth,
- execution of child leaves with matching work types,
- child landing into the parent integration target,
- bounded re-planning when discoveries stay inside the approved goal,
- use of specified actors or actor roles,
- use of specified file/resource scopes,
- validation and review requirements.

The envelope does not authorize:

- landing the root to `main`,
- policy changes,
- autonomy elevation,
- irreversible actions,
- unexpected files or resources,
- changes that alter the parent contract,
- failed validation,
- risk above the approved floor.

## Planning Horizon

The planner may propose one of three shapes:

- **Complete tree**: all children are known and concrete enough to approve as a unit.
- **Subtree plan**: the next layer is concrete, but deeper work should be planned after
  discovery or implementation.
- **Planning budget**: the parent is approved to decompose further inside limits such as
  max depth, max child count, allowed work types, and allowed resources.

This keeps dynamic planning available without requiring a human approval for every
internal expansion.

## Candidate Envelope Shape

```yaml
approval_envelope:
  id: env-T-2000
  node_id: T-2000
  approved_actions:
    - decompose_children
    - execute_children
    - land_children_to_parent
  planning:
    max_depth: 2
    max_children: 12
    allowed_work_types:
      - architecture_discovery
      - technical_design
      - technical_implementation
      - test_implementation
      - documentation
  scope:
    files:
      allow:
        - internal/runtime/**
        - internal/worktree/**
        - docs/architecture/agent-runtime.md
      require_review:
        - docs/adr/**
        - .groundwork/policies/**
      deny:
        - .env*
  validation:
    required_templates:
      - go
  escalation:
    on_unexpected_files: true
    on_public_api_change: true
    on_validation_failure: true
    on_risk_above: medium
    on_contract_change: true
```

## Approval Flow

1. Planner proposes a parent contract and envelope.
2. Policy determines required approver roles from work type, risk, scope, and action.
3. Human or authorized reviewer approves, rejects, or requests clarification.
4. Accepted children become dispatchable according to dependencies and resource locks.
5. Any child that exceeds the envelope creates an exception approval.
6. Parent integration verifies all child summaries, diffs, and validation before final
   parent or root review.

## Policy Inputs

Policy should receive both planned and actual state:

- action type,
- node depth and parent/root,
- work type,
- planned scope,
- actual touched files/resources,
- actor id and roles,
- risk class,
- reversibility,
- validation status,
- branch target,
- whether the action remains inside an approved envelope.

## Web Admin Implications

The web admin should let a reviewer approve the envelope with an explanation of:

- what the system may do,
- where it may do it,
- which roles/actors may do it,
- what will still require review,
- what validation will be required,
- which children are currently planned.

Approval queues should group exception requests under their parent envelope.

## Open Questions

- Should envelope records live only in SQLite and ticket exports, or also in policy files?
- Should envelope approval create child nodes in `todo`, or should children remain
  `backlog` until the planner marks each one ready?
- What is the minimum envelope shape needed for Phase 4 without overbuilding?
- How should envelope revocation affect active child runs?

## Candidate ADRs

- Approval envelopes as a bounded delegation mechanism.
- Planning horizons for dynamic decomposition.
- Exception approvals for scope, risk, and validation drift.

