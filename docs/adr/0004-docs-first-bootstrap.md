# ADR 0004: Docs-First Bootstrap

Status: Accepted
Implemented: Implemented

## Context

Groundwork will eventually coordinate agents, but it cannot safely self-host until the coordinator exists.

## Decision

Bootstrap with exhaustive docs and planning records before code.

## Consequences

Fresh implementation sessions can start from durable context. No app logic should be added during bootstrap.
