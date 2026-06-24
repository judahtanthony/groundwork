# Architecture Decision Records

ADRs record durable architectural decisions. They are canon once accepted, but they are
not implementation ledgers: each ADR separates decision status from implementation state.

Use this header for new ADRs:

```text
# ADR NNNN: Title

Status: Draft
Implemented: Not started
```

## Status

- `Draft`: exploratory design material. It may change freely and does not override
  accepted ADRs, contracts, or project instructions.
- `Pending Review`: proposed canon. It does not override accepted ADRs, contracts, or
  project instructions.
- `Accepted`: binding project decision.
- `Rejected`: considered and not adopted.
- `Superseded`: replaced by a later ADR; link the replacement.

## Implemented

- `Not started`: no current system behavior should be assumed from this ADR yet.
- `Partial`: some system behavior exists; the ADR body should name what exists and what
  remains.
- `Implemented`: the current code, docs, policy, or contracts reflect the decision.
- `Not applicable`: the ADR is framing or direction only and intentionally changes no
  code, policy, schema, or contract.

`Status` and `Implemented` are independent. For example, a directional ADR can be
`Accepted` and `Not applicable`, while exploratory future work should usually start as
`Draft` and `Not started` or `Partial`.

Use ADRs for both exploration and accepted decisions. Move a draft to `Pending Review`
when it is ready to be reviewed as proposed canon; move it to `Accepted` only when the
decision is binding.

Recent accepted ADRs:

- ADR 0051 records async agent handoff and durable ticket-attached decision records.
- ADR 0052 records consequential decisions as ordinary policy-routed work nodes.
- ADR 0053 records filesystem-authoritative durable state with SQLite as projection.
