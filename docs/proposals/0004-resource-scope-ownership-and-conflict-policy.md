# Proposal 0004: Resource Scope, Ownership, And Conflict Policy

Status: Draft

## Goal

Prevent agents from stepping on each other by making planned resource scope a first-class
part of nodes, policy, scheduling, and landing review.

## Problem

Dependency edges answer ordering, but not contention. Two independent nodes can both edit
the same file, change the same interface, update adjacent tests, or modify the same
product surface. The planner should identify dependencies, but it cannot reliably ensure
there are no file or semantic conflicts without explicit scope and ownership data.

## Proposal

Each planned node should be able to declare resource scope. The scheduler uses planned
scope to avoid conflicting parallel runs. Landing compares actual diff scope against
planned scope. Policy uses both planned and actual scope to require owner review,
escalation, or auto-approval.

## Resource Scope Shape

```yaml
scope:
  files:
    allow:
      - internal/runtime/**
      - internal/worktree/**
    require_review:
      - docs/adr/**
      - .groundwork/policies/**
      - .github/workflows/**
    deny:
      - .env*
      - "**/*secret*"
  symbols:
    allow:
      - runtime.Runtime
      - runtime.Spec
    require_review:
      - policy.AuthorizeClaim
  resources:
    reserve:
      - codex_runtime_adapter
      - run_event_contract
  owners:
    - runtime
```

The initial implementation can start with file globs. Symbols and named resources can be
added after the file-scope path proves useful.

## Scope Semantics

- `allow`: actor may touch this scope if other gates pass.
- `require_review`: actor may touch this scope, but landing or parent integration needs
  a matching role/owner approval.
- `deny`: actor may not touch this scope under the current envelope; touching it creates
  an exception.
- `reserve`: named non-file resource that should not be modified in parallel by another
  active node without explicit approval.

## Scheduling Rules

The scheduler should consider global active scopes:

- Two nodes with overlapping exclusive file scopes should not run in parallel unless a
  policy rule allows shared access.
- A root node can reserve broad resources before its children start.
- If a new root conflicts with an active root, the planner should either add a dependency,
  narrow the scope, or request approval to proceed in parallel.
- Advisory conflicts can warn without blocking in early phases.

## Landing Rules

Landing should compare actual touched files against planned scope:

- In-scope, low-risk, validation-passed changes may qualify for batch or auto landing to
  a parent branch.
- Unexpected files require exception review.
- Files under owner-controlled paths require the relevant role.
- Policy-sensitive files, secrets, CI/CD, release, auth, and migrations should raise risk.

## CODEOWNERS-Style Policy

Policy matching should support scope and ownership:

```yaml
rules:
  - id: ci_config_requires_staff
    when:
      action_types:
        - land_to_parent
        - land_to_main
      files:
        - .github/workflows/**
        - scripts/deploy/**
    require_roles:
      - staff_engineer

  - id: docs_low_risk_child_land
    when:
      action_types:
        - land_to_parent
      work_types:
        - documentation
      risk_class_at_most: low
      actual_scope_within_planned: true
      validations_passed: true
    outcome: allow
```

## Dynamic Risk Inputs

Risk should be recalculated from what actually happened:

- actual touched files,
- diff size,
- public API or schema changes,
- validation failures,
- actor trust and track record,
- whether the run stayed in scope,
- whether clarification was required,
- whether sibling or cross-root conflicts exist.

## Web Admin Implications

The web admin should show scope visibly:

- planned scope,
- actual touched files,
- owner-required paths,
- conflict warnings,
- exception reasons,
- policy rule that triggered the gate.

This makes approval explainable instead of a generic "needs human" state.

## Open Questions

- How strict should scope enforcement be before the real Codex runtime exists?
- Should file-scope locks block scheduling, or only warn, during early adoption?
- How should generated files and formatting-only changes be classified?
- Should resource scopes be stored in ticket exports, parent contracts, or both?

## Candidate ADRs

- File/resource scope as a first-class node contract.
- CODEOWNERS-style policy matching for Groundwork.
- Actual-diff scope comparison as a landing gate input.

