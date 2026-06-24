# Architecture Decision Records

ADRs record durable architectural decisions. They are canon once accepted, but they are
not implementation ledgers: each ADR separates decision status from implementation state.

Use this header for new ADRs:

```text
# ADR NNNN: Title

Status: Pending Review
Implemented: Not started
```

## Status

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
`Accepted` and `Not applicable`, while a proposal migrated from `docs/proposals/` should
usually start as `Pending Review` and `Not started` or `Partial`.
