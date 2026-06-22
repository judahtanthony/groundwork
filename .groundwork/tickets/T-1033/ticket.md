---
id: T-1033
kind: ticket
node_type: leaf
work_type: technical_implementation
title: Scope CLI->coordinator routing to the project root (cross-project mutation footgun)
status: backlog
assignee: null
requested_actor: null
priority: 0.7
labels:
    - cli-ux
    - coordinator
parent: T-1022
depends_on: []
created_at: "2026-06-22T19:33:45Z"
updated_at: "2026-06-22T20:53:56Z"
---

## Problem

gw ticket mutations (create/transition/etc.) auto-route to a coordinator on the default 127.0.0.1:4500 whenever one is running, without checking that the coordinator serves the same project root the CLI was invoked in. Consequence: running 'make smoke' (or any gw command in another repo/temp dir) while a 'gw server' is up silently writes into the SERVER's project, not the CWD's. Reproduced: with gw server up on :4500 (this repo), 'make smoke' created its 'Build store' tickets here (T-1033/T-1034) and then failed looking for T-0001 in its own throwaway repo. Fix options: project-scoped coordinator address (per-project socket/port), or have the CLI verify the coordinator's root matches the discovered project root before routing (else fall back to store-direct or error), and/or have smoke pin a distinct addr/env. Surfaced landing T-1023.

## Acceptance Criteria

- openTicketStore only routes to a coordinator whose project root matches the CLI's discovered root (e.g. /healthz reports the root); otherwise falls back to store-direct or errors clearly
- Test/ephemeral apps are isolated: scaffolded/temp configs use a non-:4500 port (or 127.0.0.1:0) so pre-server CLI steps in make smoke go store-direct and never hit a running coordinator
- Per-project Unix-domain socket under .groundwork/ is recorded as the preferred long-term design (collision-proof, no port management), even if (a) lands first
