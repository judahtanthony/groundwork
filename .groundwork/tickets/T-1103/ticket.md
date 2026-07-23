---
id: T-1103
kind: ticket
node_type: null
work_type: technical_implementation
title: 'Envelope supersession is broken: approving a new envelope while one is active silently skips activation'
status: backlog
assignee: null
requested_actor: null
priority: null
labels:
    - workflow
parent: T-1074
depends_on: []
created_at: "2026-07-22T23:11:33Z"
updated_at: "2026-07-22T23:11:33Z"
---

## Problem

activateEnvelope (internal/server/envelope_service.go:80-84) returns early via ensureIntegrationBranch whenever GetActiveEnvelopeForNode returns any active envelope. This idempotency guard (intended for retry safety after a partial activation) conflates 'this same envelope already activated (retry)' with 'a DIFFERENT envelope is already active (supersession)'. Result: proposing + approving a widened/replacement envelope for a node that already has an active one records the approval but never materializes the new draft — the old envelope stays active with stale scope, silently. Discovered dogfooding T-1036: widening the web envelope to add docs/contracts/** (approval A-0003) was approved but ENV-0001 remained active unchanged. ADR 0054 lists supersede in the lifecycle but activation never calls SupersedeEnvelope. Current manual workaround: revoke the active envelope, then propose+approve the replacement. OUTSIDE the T-1036 web envelope; platform/workflow work.

## Acceptance Criteria

- Approving an approve_envelope proposal for a node that already has an active envelope supersedes the old envelope and activates the new draft (old -> superseded, new -> active, sidecar + mirror updated)
- True idempotent retry (re-activating the SAME approved draft) still no-ops without allocating a duplicate envelope id or orphaning a sidecar
- A regression test covers propose+approve of a second envelope while one is active, asserting the new scope is active and the old is superseded
