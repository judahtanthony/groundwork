# ADR 0005: Localhost Single-User V1

Status: Accepted

## Context

The primary v1 audience is a solo developer running locally. LAN/team mode would require authentication, authorization, CSRF, network safety, and more operations work.

## Decision

Make v1 localhost-only and single-user.

## Consequences

The server can focus on coordination, dashboard, approvals, and recovery without a full multi-user security model.

