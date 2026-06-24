# ADR 0029: Actor Identity Model

Status: Accepted
Implemented: Implemented

## Context

[ADR 0023](0023-actors-work-types-and-policy-routing.md) shipped the actor
registry with a closed `type` enum and flat, free-form dot-segmented ids, and
explicitly marked the identity scheme **provisional**, to be settled early in
Phase 2 "when actor-aware matching first binds code to the model." Phase 2's gate
engine ([ADR 0028](0028-gate-evaluation-engine.md)) is that binding point, so the
model must be resolved now.

Two generalizations were anticipated: **class vs instance** (route to a fungible
capability set; audit the unique entity) and **tiered identity** (a dotted,
prefix-matchable namespace where a policy can start coarse and tighten). A
tempting concrete scheme is a fixed three-segment path —
`{authority}.{role}.{identity}`, e.g. authority `{human, agent}` first, broad
role/capability second, full identity third. That is enough for a one-person,
one-agent bootstrap, but it **bakes role into a fixed identity segment**, which
is exactly the corner to avoid: identities specialize and evolve over time, and
capability is best expressed as several independent properties, not one
mutually-exclusive group. Entangling the two would force re-pointing an
identity's path every time its role changes.

## Decision

Model an actor along **two orthogonal axes**: a hierarchical **identity path** and
a parameterized **capability object**. Policy matches a parameterized combination
of both.

### Identity path — authority and grouping only, never capability

- The actor `id` is a **dotted path of authority/grouping tiers, prefix-matchable
  at any depth**, bottoming out in a unique instance. Tier 1 is the coarse
  **authority class** (e.g. `human`, `ai`) — sufficient for the bootstrap and for
  coarse rules ("a `human` must approve"). Deeper tiers are **grouping toward a
  unique instance** and have **conventional, not schema-enforced, meaning**; the
  depth at which a rule matches is the fungibility boundary.
- The id root is **not validated against `type`**. Doing so would require a fixed
  root vocabulary, contradicting the conventional-tier principle and the
  bootstrap registry, which uses `ai.*` ids with `ai_agent`/`ai_judge` types.
  `type` remains its own independent matchable dimension.
- The path expresses **who / authority / grouping**. It deliberately does **not**
  encode role, work type, or capability. A second segment that happens to read
  like a role (e.g. the bootstrap's `human.owner`) is a stable grouping label,
  **not** the authoritative role — policy must read role from the capability
  object, not by parsing a segment. This is what keeps identity orthogonal:
  an actor can specialize without rewriting its identity, and a re-org changes
  groupings without redefining capabilities.
- Variable depth is legal. `agent`, `agent.codex`, `agent.codex.default` are all
  valid; a rule matching `agent` covers every agent instance beneath it.

### Capability — a parameterized property object, not a tier

- Capabilities are an **open set of independent properties**: `work_types`,
  `roles`, review/approve authority, `runtime`/`model`, `sandbox`, allowed/denied
  file scopes, risk `limits`, and (forward-compatible) tools/skills/MCPs. This is
  the existing `internal/actor.Capabilities`, extended additively. Properties are
  matched independently (set membership / predicate), never collapsed into one
  exclusive group.

### Class vs instance

- **Routing matches a class**: an identity prefix plus required capability
  predicates ("an `agent` that can claim `documentation` at ≤ `medium` risk").
  `requested_actor` is a class/prefix request and a hint only.
- **Audit records the instance**: the resolved concrete registry actor, persisted
  as run `actor_id` + `actor_snapshot_json` and as a node's resolved `assignee`.
- M2 has **one instance per class**, so resolution is "prefix request + capability
  predicate → the matching registry actor." Pools of interchangeable instances
  per class are the documented scaling shape but are **not built in M2**.

### Policy matcher

A rule is a parameterized predicate (see [ADR 0028](0028-gate-evaluation-engine.md)):
identity-prefix (`actor_ids`) AND `actor_types` AND `roles` AND `work_types` AND
capability predicates AND action/files/risk/reversibility. Absent conditions are
wildcards. Identity prefix and capability are separate dimensions of the same AND,
never conflated.

### Granularity — few actors, specialization via work types and SOPs

Capability has multiple factors, which raises a real risk: if a new actor were
minted per capability combination, actors would scale combinatorially into
near-duplicates. We avoid this by keeping three things distinct:

- **Actors stay few and coarse.** An actor (class) is distinguished only by what
  is genuinely expensive or distinct — runtime, model, trust tier, sandbox,
  provider — not by fine capability tuples. The bootstrap's single generalized
  agent is the default; split into more actors only when a real underlying
  difference exists (a different model, a higher trust level, another provider),
  never to encode a narrow skill.
- **Specialization lives in `work_type` + SOP, not in the actor.** Per
  [ADR 0023](0023-actors-work-types-and-policy-routing.md) and
  [ADR 0011](0011-progressive-planning-autonomy-via-sops.md), the same
  generalized agent *becomes* a planner, researcher, or coder by being routed to
  a `planning`/`research`/`implementation` work type whose **SOP supplies the
  instructions, tools, context, and validations**. Specialization is data (SOPs
  keyed by work type), behavioral, and cheap to add — not a new identity.
- **Capabilities are an authorization *filter*, not a behavior *generator*.** The
  factors (`work_types`, `roles`, max risk, file scopes) only narrow *which of
  the few actors may claim a node*. A filter with N factors constrains matching;
  it does not enumerate the product space. Routing gets more precise without
  multiplying actors, so there is no combinatorial blow-up: a few actors plus a
  few first-match policy rules, never an actor per combination.

This makes the generic→specialized progression a **trust/SOP gradient, not an
actor-proliferation gradient**. A generalized agent earns autonomy *per
(work_type × actor)* track record via policy autonomy levels (a human act in v1,
[ADR 0011](0011-progressive-planning-autonomy-via-sops.md)). A **dedicated
specialist instance** (a deeper identity tier) is carved out only when track
record plus a clear quality advantage justify trusting it autonomously on a
narrow problem — the earned exception, not the default. Routing stays simple
(`work_type` → policy rule → coarse actor class → instance); the SOP does the
specialization.

## Consequences

The `type` enum is relaxed to "coarsest tier, consistent with the id root"; ids
become prefix-matchable; capabilities stay a property set. `internal/actor` gains
a prefix matcher and a class→instance resolver; the registry schema is unchanged
for the bootstrap (`human.owner`, `ai.codex.default`) and forward-compatible. We
explicitly reject binding role to a fixed identity segment. The model stays simple
for the solo-developer default (two actors, flat-looking ids) while the
two-axis + prefix design leaves room for evolving specialists, instance pools,
multiple runtimes/providers, and tightening rules — without a future schema break.
The provisional section of [ADR 0023](0023-actors-work-types-and-policy-routing.md)
is now resolved by this ADR.
