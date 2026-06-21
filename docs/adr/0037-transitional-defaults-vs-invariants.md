# ADR 0037: Transitional Defaults Versus Architectural Invariants

Status: Accepted

## Context

ADR 0036 ratifies a vision in which human authority is a retractable position on a trust
gradient, not hardcoded structure. But the current docs frame conservative v1 gates as
permanent **commitments**: AGENTS.md lists "human approval required for landing and
decomposition" under *Design Commitments To Preserve*, and ADR 0014 states irreversible
actions are "never auto-approved, regardless of score." Because this repo treats committed
docs as the source of truth, these read as bedrock. To move authority along the gradient
without re-litigating the foundations each time, the two kinds of decision must be
distinguished explicitly.

## Decision

Classify every safety-relevant decision as either a **transitional default** (a
conservative *current setting* that may loosen as SOPs, validation, context, and trust
mature) or an **architectural invariant** (a property preserved at *every* autonomy level,
including full autonomy).

**Transitional defaults** (loosenable; conservative today, not permanent):

- Human approval required to land to `main` (ADR 0006).
- Human approval required to accept a decomposition (ADR 0011).
- Irreversible actions routed to a human (ADR 0014; mechanism amended by ADR 0038).
- Autonomy elevation performed by a human (ADR 0011 "Policy Learning"; amended by ADR 0038).
- Single-user, localhost-only operation (ADR 0005).

**Architectural invariants** (preserved across all autonomy levels):

- **Auditability** — every state-changing action appends an audit record in the same
  transaction (`sqlite-schema.md`). Autonomy may remove the human, never the record.
- **Reversibility as a tracked gate input** — reversibility remains classified and shown
  on every gated action (ADR 0014). What loosens (ADR 0038) is the *consequence* of
  irreversibility, never whether it is evaluated and surfaced.
- **Canon-as-memory** — durable design is distilled and ratified, not lost (ADR 0013).
- **Deterministic, file-authoritative durable state** — deterministic export (ADR 0020)
  and the committed-vs-runtime tiering (ADR 0012).
- **Single serialized coordinator** for claim arbitration and canon writes (ADR 0026).
- **Default-deny authorization** — no claim or gated action proceeds without a matching
  rule (ADR 0028). Loosening adds rules; it never removes the default-deny floor.

A decision's classification is part of its ADR. New safety-relevant decisions must state
which category they fall in.

## Consequences

The "humans hold everything" posture is recorded as the current *default configuration*,
not the architecture. AGENTS.md's *Design Commitments* section is split accordingly: the
human-gate items move under transitional defaults pointing here, while the invariants
remain commitments. This ADR changes framing only — no code, policy, or schema — and gives
ADR 0038 and later autonomy work a stable boundary to loosen *up to* but not *past*.
