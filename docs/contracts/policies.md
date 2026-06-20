# Policy Contracts

Policies live under `.groundwork/policies/` and are committed by default. Rules carry stable IDs and are evaluated top-down; the first matching rule decides. The dashboard surfaces these IDs (for example `R-01`) for auditability. Policies may match actors from `.groundwork/actors.yaml` by id, type, role, runtime, and capability.

Every rule carries one uniform `when:` predicate; all match conditions (files, actor ids/types/roles, work types, action types, risk class, reversibility, commands) live under it (normalized per [ADR 0028](../adr/0028-gate-evaluation-engine.md)). Conditions present in `when:` must all hold (logical AND); absent conditions are wildcards.

## Trust Policy Example

```yaml
schema: groundwork_trust_policy/v1
auto_approve:
  - id: internal_docs
    description: Allow documentation-only changes to internal agent guidance.
    when:
      files:
        - AGENTS.md
        - MEMORY.md
        - ".groundwork/**/*.md"
      change_type: documentation
      max_diff_lines: 200
  - id: safe_tests
    when:
      command_regex: "^(go test|npm test|pytest|make test)"
      cwd_within_workspace: true
      network: false
  - id: docs_ai_judge
    when:
      work_types: [documentation]
      risk_class: low
    review_allowed_by:
      actor_ids: [ai.docs_judge]
require_human:
  - id: secrets
    when:
      files:
        - "**/.env*"
        - "**/*secret*"
  - id: destructive_commands
    when:
      command_categories: [destructive]
  - id: landing_to_main_v1
    when:
      action_types: [land_to_main]
  - id: decomposition_v1
    when:
      action_types: [decompose]
allow_claim:
  - id: default_codex_medium_risk
    when:
      actor_ids: [ai.codex.default]
      work_types: [technical_design, technical_implementation, test_implementation, documentation]
      risk_class_at_most: medium
    actions: [execute, decompose]
  - id: billing_human_only
    when:
      files: ["billing/**", "payments/**"]
      actor_types: [human]
```

Gated actions (`execute`, `land_to_main`, `decompose`) carry an autonomy level that can be loosened as work-type SOPs, context, and validations mature. Elevation is a human act in v1; see `docs/architecture/trust-and-approvals.md` and `docs/adr/0011-progressive-planning-autonomy-via-sops.md`.

Actor-aware policy matching is the smallest v1 authorization model. It is local and declarative: rules match action, work type, file scope, risk class, reversibility, actor id, actor type, and actor roles/capabilities. The result may allow a claim, require an approval from a role, deny an edit, or select a reviewer. This supports the single-developer default while preserving a path to multiple human roles and task-specific AI actors.

Gates also key off **reversibility**: an action classified irreversible (non-reversible migration, external/production state, destructive command, credential access) is forced to `critical` and stays human-required regardless of risk score or autonomy level (see `docs/adr/0014-reversibility-as-a-gate-input.md`). A rule may set or assert `reversible: false` to force this.

## Autonomy Levels Example

```yaml
schema: groundwork_autonomy_policy/v1
actions:
  execute:
    default: require_human       # require_human | reviewer (phase 2) | auto
  land_to_main:
    default: require_human
  decompose:
    default: require_human
    by_work_type:
      docs:
        level: auto              # earned via SOP + validations; elevated by a human
        requires:
          sop: sops/docs/
          validations: [documentation]
```

## Validation Policy Example

```yaml
schema: groundwork_validation_policy/v1
templates:
  documentation:
    match:
      files: ["**/*.md", "AGENTS.md", "MEMORY.md"]
    required: []
    landing_risk_floor: low
  go:
    match:
      files: ["**/*.go"]
    required:
      - name: go_tests
        command: "go test ./..."
```
