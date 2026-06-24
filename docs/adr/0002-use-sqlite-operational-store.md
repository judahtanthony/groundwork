# ADR 0002: Use SQLite As Operational Store

Status: Accepted
Implemented: Implemented

## Context

Using Git and files as the live lock manager would complicate claims, leases, approvals, and pause/resume.

## Decision

Use `.groundwork/state.sqlite` as the local operational store during runtime. Ignore it by default and export durable application state to readable files.

## Consequences

Active coordination is simpler and transactional. Recovery must reconcile SQLite, durable exports, worktrees, and run logs.
