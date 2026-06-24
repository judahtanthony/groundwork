# Proposal 0006: Role-Aware Actors And Local Identity

Status: Draft

## Goal

Support a simple MVP where one engineer acts as every approver while preserving a clean
path to multiple human roles, specialized AI agents, and user-aware CLI and web-admin
surfaces.

## Problem

The current MVP is single-user and local, but the system will need to distinguish roles:
a product owner may approve top-level product plans, a staff engineer may approve CI/CD
or architecture-sensitive changes, and specialized AI agents may need different tools,
permissions, environments, or work-type capabilities.

The CLI and web admin also need a way to answer "my work" without requiring users to pass
an actor id on every command.

## Actor ID Convention

Use readable, namespaced actor ids:

```text
human.fe-engineer.john-doe
human.staff-engineer.alex
human.pm.sara

ai.fe-engineer.codex
ai.test-engineer.codex
ai.docs-writer.codex
ai.release-engineer.codex
```

The second segment is a useful role namespace for routing and display. It should not be
the only source of truth for authority.

## Structured Actor Fields

Policy should match structured fields as well as ids:

```yaml
actors:
  - id: human.fe-engineer.john-doe
    type: human
    display_name: John Doe
    roles:
      - fe_engineer

  - id: human.staff-engineer.alex
    type: human
    display_name: Alex
    roles:
      - staff_engineer
      - ci_owner

  - id: ai.fe-engineer.codex
    type: ai_agent
    display_name: Codex Frontend Engineer
    runtime: codex
    roles:
      - fe_engineer
      - implementer
    capabilities:
      work_types:
        - frontend_implementation
        - visual_design_followup
      file_scopes:
        - app/**
        - components/**
```

Keep these concepts separate:

- actor id: stable identity,
- roles: approval and authority semantics,
- capabilities: what this actor can do,
- preferences: what work a local user wants to see,
- runtime permissions: what an agent process may access.

## Policy Matching

The policy matcher should support:

- actor ids,
- actor id prefixes,
- actor roles,
- actor type,
- capabilities,
- work type,
- file/resource scope,
- required approver role.

Example:

```yaml
allow_claim:
  - id: frontend_agent_can_claim_frontend
    when:
      actor_roles:
        - fe_engineer
      work_types:
        - frontend_implementation
      files:
        - app/**
        - components/**
      risk_class_at_most: medium
    actions:
      - execute
```

## Required Approval Roles

Approval records should say which role was required and which actor satisfied it:

```yaml
approval:
  action: land_to_main
  required_roles:
    - staff_engineer
  satisfied_by: human.owner
  reason: touched .github/workflows/test.yml
  policy_rule: ci_config_requires_staff
```

For MVP, `human.owner` may have all roles. Later, those roles can map to different humans
without changing the workflow model.

## Local Identity

Add layered identity config for "my work":

1. Committed project actors: `.groundwork/actors.yaml`
2. Untracked project-local identity: `.groundwork/local.yaml`
3. Global user identity: `~/.config/groundwork/config.yaml` or equivalent

Example:

```yaml
default_actor: human.fe-engineer.john-doe
preferred_roles:
  - fe_engineer
```

Commands can then default naturally:

```text
gw mine
gw next --mine
gw ticket list --mine
gw approval list --mine
gw ticket claim T-1234
```

"Mine" can mean:

- assigned to my actor,
- approval requires a role my actor satisfies,
- review requested from my actor or role,
- work type matches my preferred roles or capabilities,
- work claimed or started by my actor.

## Web Admin Implications

Without authentication, the MVP web admin can still display:

```text
Acting as: John Doe
Roles: FE Engineer
```

Later authentication should bind a logged-in user to one or more actor ids. The policy
engine should receive an actor either way.

## Open Questions

- Should local identity live under `.groundwork/local.yaml` or a home-directory config by
  default?
- Should actor id segment conventions be validated or only recommended?
- Can one UI session switch actors for testing or role simulation?
- Should specialized AI agents be separate actors or one actor with selectable profiles?

## Candidate ADRs

- Actor id namespace convention.
- Explicit actor roles and capabilities as policy inputs.
- Local identity resolution for CLI and web admin.
