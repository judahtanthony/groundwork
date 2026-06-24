# ADR 0011: Progressive Planning Autonomy Via SOPs

Status: Accepted
Implemented: Partial

> Implemented today: SOP files, action-level autonomy defaults, `decompose` gates,
> and policy-based claim authorization. Not implemented: earned autonomy,
> trust-tier evaluation, or self/elevated policy changes.
>
> Amended by ADR 0038: trust elevation becomes a first-class gated action
> (`amend_policy` / `elevate_autonomy`) rather than a human-only act — human-required by
> default, but expressible and delegable. The human requirements here are loosenable policy
> defaults per ADR 0037.

## Context

ADR 0006 models landing as a policy gate so autonomy can be enabled later. Dynamic decomposition (ADR 0009) introduces a new high-leverage agent action — planning — that is human-gated in v1. Planning should be able to loosen over time the same way leaf execution and landing do, rather than staying permanently manual.

## Decision

Treat all high-leverage agent actions uniformly — leaf `execute`, `land_to_main`, and the new `decompose` — as capabilities with a risk class and an approval requirement that can be progressively loosened. Loosening is earned as work-type **SOPs** (standard operating procedures), updatable work-type **context**, and defined **validations** mature, giving each action class an autonomy level that can move from human-required toward policy/auto.

SOPs and work-type context are committed, durable artifacts under `.groundwork/sops/<work-type>/`, separate from the single global `.groundwork/WORKFLOW.md`. Trust **elevation is itself a human act** in v1: Groundwork may suggest loosening after repeated clean approvals but never self-elevates (extending the Policy Learning rule in `trust-and-approvals.md`).

## Consequences

Trust policy gains a `decompose` action type and a per-action autonomy-level notion. A new committed SOP directory is added to the file layout. There is one consistent ladder from human-gated to autonomous for execution, landing, and planning, keeping v1 conservative while preserving the path to Symphony-style autonomy.
