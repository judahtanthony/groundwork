# ADR 0006: Human Landing Gate In V1

Status: Accepted
Implemented: Implemented

> The human landing gate is reclassified as a **transitional default** (loosenable, not a
> permanent commitment) by ADR 0037.

## Context

The long-term goal is high autonomy, but early trust and validation policies will be immature.

## Decision

Require human approval before landing code to `main` in v1. Model landing as a policy gate so future autonomous landing can be enabled safely.

## Consequences

V1 is safer while preserving the path to Symphony-style zero-human-code operation.
