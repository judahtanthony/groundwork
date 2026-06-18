# ADR 0005: Localhost Single-User V1

Status: Accepted

## Context

The primary v1 audience is a solo developer running locally. LAN/team mode would require authentication, authorization, CSRF, network safety, and more operations work.

## Decision

Make v1 localhost-only and single-user.

## Consequences

The server can focus on coordination, dashboard, approvals, and recovery without a full multi-user security model.

"Single-user" means a single **human operator** — no accounts, authentication, or permission service in v1. It does **not** mean a single actor: the work graph is acted on by multiple actors, including the human owner and one or more AI actors (see [ADR 0023](0023-actors-work-types-and-policy-routing.md)). Multiple *human* roles, and the authentication/authorization they require, are deferred to the post-v1 remote/LAN mode (see the roadmap).

