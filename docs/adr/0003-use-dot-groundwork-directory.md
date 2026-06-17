# ADR 0003: Use `.groundwork/`

Status: Accepted

## Context

The system needs colocated state without forcing a monorepo layout or colliding with existing agent config.

## Decision

Use `.groundwork/` as the managed-project dot directory.

## Consequences

Adoption is low-friction and repo-local. Existing `AGENTS.md`, `.codex/`, or other agent files remain separate.

