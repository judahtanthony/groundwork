---
id: E-0012
kind: epic
node_type: composite
work_type: null
title: Activate envelope and escalation enforcement with the runtime diff
status: done
assignee: null
requested_actor: null
priority: 0.5
labels:
    - phase-6
    - bounded-autonomy
parent: T-1071
depends_on: []
created_at: "2026-06-28T22:00:00Z"
updated_at: "2026-06-28T22:00:00Z"
---

## Problem

Phase 5 built the envelope model, the AND-composition gate, and exception approvals, but
left two seams inert because there was no runtime diff: the scheduler still claims AI work
trust-only (not envelope-aware), and the envelope file-scope / escalation triggers are
recorded but never enforced. Phase 6's runtime supplies the diff, so this epic activates
both: envelope-aware claim routing and diff-fed escalation enforcement.

## Acceptance Criteria

- AI claims are authorized by trust AND the active envelope; boundary crossings raise
  human exception approvals (T-1090).
- Envelope file-scope and the five escalation triggers are enforced against the run diff,
  preserving the human-gated invariants (T-1091).
