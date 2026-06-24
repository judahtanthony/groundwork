# ADR 0001: Use Go

Status: Accepted
Implemented: Implemented

## Context

Groundwork needs a portable CLI, local coordinator, SQLite access, subprocess supervision, HTTP/SSE, and filesystem/Git operations.

## Decision

Implement v1 in Go.

## Consequences

Go supports a small portable binary and low runtime overhead. A richer TypeScript frontend can be added later if needed, but v1 should avoid requiring Node for normal use.
