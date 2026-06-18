# Conventions

## Documentation

- Keep root docs short and navigable.
- Put long-lived product intent in `docs/product/`.
- Put system design in `docs/architecture/`.
- Put public surfaces and data shapes in `docs/contracts/`.
- Put major decisions in `docs/adr/`.
- Put UI/visual design reference in `docs/design/` (reference only, not app code).
- Keep `docs/reference/` concise and easy for agents to query.

## Planning

- Use the uniform work tree; `kind` is an advisory label, the structural fact is leaf vs composite.
- Use `work_type` for SOP, actor-routing, validation, and policy semantics; do not add organization-specific SDLC phases to the status model.
- Triage nodes on claim: leaf nodes execute; composite nodes decompose just-in-time as a reviewable proposal.
- Prefer a complete parent contract so children run in parallel; otherwise add dependency edges to serialize.
- Keep leaf nodes to one verifiable change.
- Use escalation to propagate revisions up the tree.
- Capture work-type operating procedures as SOPs under `.groundwork/sops/`.
- Capture local human and AI actors in `.groundwork/actors.yaml`; requested actors are routing hints, not authorization.
- Link nodes to specs or ADRs where relevant.

## Future Code

- Prefer Go standard library when clean enough.
- Prefer small, mature, open-source dependencies.
- Avoid mandatory external services in v1.
- Keep package boundaries aligned with the architecture map.

## State

- Do not commit runtime state by default.
- Do commit durable docs, policies, workflow, ticket exports, and code.
- Generated views are not source of truth.
