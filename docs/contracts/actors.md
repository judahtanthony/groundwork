# Actor Registry Contract

`.groundwork/actors.yaml` is committed durable configuration for local actors. It is intentionally small in v1 and forward-compatible with richer collaboration later.

## MVP Shape

```yaml
schema: groundwork_actors/v1

actors:
  - id: human.owner
    type: human
    display_name: Owner
    roles: [owner]
    capabilities:
      work_types: ["*"]
      approve: ["*"]
      review: ["*"]

  - id: ai.codex.default
    type: ai_agent
    display_name: Codex Default
    runtime: codex
    model: default
    roles: [implementer]
    capabilities:
      work_types:
        - technical_design
        - technical_implementation
        - test_implementation
        - documentation
      review:
        - documentation
    limits:
      max_risk_class: medium
    sandbox: workspace-write
```

## Fields

- `schema`: must be `groundwork_actors/v1`.
- `actors`: ordered actor definitions. Actor ids must be stable and unique.
- `id`: stable id used in tickets, runs, approvals, audit events, and policies.
- `type`: `human`, `ai_agent`, `ai_judge`, or `tool`.
- `display_name`: human-readable name.
- `roles`: organization-defined role labels used by policy.
- `runtime`: AI runtime adapter name, for example `codex`.
- `model`: model or model alias used by the runtime.
- `capabilities`: coarse routing and authorization claims. MVP supports `work_types`, `approve`, and `review`; later versions may add tools, MCPs, skills, domains, languages, and instructions.
- `limits`: optional risk or scope limits.
- `sandbox`: default sandbox posture for AI runs.

Unknown fields should warn but not fail, preserving forward compatibility.

## Default Semantics

`human.owner` is the local fallback approver for v1. `ai.codex.default` is the default AI actor when no ticket requests a specific actor and policy allows an AI claim.

Actor definitions are not copied into SQLite as rows. Runs persist `actor_id` and `actor_snapshot_json`; approvals persist requesting and deciding actor ids. The registry remains the current routing source, while snapshots preserve historical audit.

## Policy Matching

Policies may match:

- `actor_ids`,
- `actor_types`,
- `roles`,
- `runtime`,
- `work_types`,
- action type,
- files,
- risk class,
- reversibility.

A requested actor on a ticket is never sufficient by itself. Policy must still allow the claim, review, approval, edit, or landing action.

