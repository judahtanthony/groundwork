# Proposal 0007: Full-SDLC Work Modeling

Status: Draft

## Goal

Use Groundwork's uniform work-node model for the full software-development lifecycle, not
only code implementation. Product discovery, architecture discovery, visual design,
technical design, implementation, tests, docs, review, and release validation should all
be representable as nodes with work-type-specific SOPs, policies, actors, and validation.

## Problem

Software quality depends on more than writing code. Some work must discover requirements,
compare architectures, design UI, update docs, create tests, or validate releases. If
Groundwork only models implementation, important dependencies and review gates remain
outside the system.

## Proposal

Keep status generic and use `work_type` for SDLC semantics. Work type drives:

- SOP selection,
- actor routing,
- validation templates,
- approval policy,
- autonomy level,
- context assembly,
- output expectations.

## Candidate Work Types

```text
product_discovery
product_spec
architecture_discovery
technical_design
visual_design
frontend_implementation
backend_implementation
technical_implementation
test_implementation
documentation
review
release_validation
```

This list should remain project-configurable. Groundwork should provide examples, not a
hardcoded methodology.

## Example Graph

```text
product_discovery
  -> product_spec
      -> architecture_discovery
          -> technical_design
              -> backend_implementation
              -> frontend_implementation
              -> test_implementation
          -> visual_design
              -> frontend_implementation
      -> documentation
      -> release_validation
```

Dependencies express ordering. Parent contracts and resource scopes express boundaries.
Policy expresses who can approve each kind of work.

## Work-Type Outputs

Different work types should produce different outputs:

- product discovery: problem statement, users, constraints, open questions,
- product spec: acceptance criteria, non-goals, testable outcomes,
- architecture discovery: options, tradeoffs, recommendation,
- technical design: contracts, package boundaries, migration plan,
- visual design: screens, flows, accessibility constraints, assets,
- implementation: code diff, tests, summary, risks,
- test implementation: coverage target and commands,
- documentation: canon updates,
- release validation: end-to-end results and release notes.

## SOP Requirements

Each mature work type should have an SOP under `.groundwork/sops/<work-type>/` defining:

- scope,
- inputs,
- expected outputs,
- review checklist,
- validation,
- escalation triggers,
- autonomy prerequisites.

## Approval And Autonomy

Work types should mature independently. Documentation may earn low-risk autonomy before
CI/CD, migrations, auth, or release work. A work type can loosen gates only when SOPs,
validation, context, and track record support it.

## Web Admin Implications

The web admin should make the SDLC graph readable without turning into a generic project
management board. Useful views:

- root feature progress by SDLC stage,
- ready and blocked nodes,
- approvals by required role,
- design decisions awaiting review,
- implementation branches,
- release validation status.

## Open Questions

- Which work types should be built into scaffolding versus left to project config?
- Should visual design artifacts remain docs-only until the web surface phase?
- How should product-owner approvals interact with engineering approvals?
- Should work-type maturity be tracked as policy state or derived from SOP and validation
  history?

## Candidate ADRs

- Work type as the SDLC extension point.
- SOP maturity as an autonomy prerequisite.
- Release validation as a first-class work type.

