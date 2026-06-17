# ADR 0007: Runtime State Is Not Committed By Default

Status: Accepted

## Context

SQLite, transcripts, raw logs, approvals, and worktrees can be volatile, large, or sensitive.

## Decision

Ignore runtime state by default. Commit durable docs, workflow, policies, ticket exports, and code.

## Consequences

Repositories stay cleaner and safer. Groundwork must export enough durable application state and recover stable operation after crashes.

