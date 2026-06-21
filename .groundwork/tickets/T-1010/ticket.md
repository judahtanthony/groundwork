---
id: T-1010
kind: ticket
node_type: leaf
work_type: technical_implementation
title: 'Reversibility: highest-bar condition, not pre-policy short-circuit'
status: backlog
assignee: null
requested_actor: null
priority: null
labels: []
parent: T-1009
depends_on: []
created_at: "2026-06-21T20:08:43Z"
updated_at: "2026-06-21T20:08:43Z"
---

## Problem

Per ADR 0038: gate.go evaluates reversibility through rules as a very-high-bar condition (still always computed/surfaced, an ADR 0037 invariant), not a return-before-policy short-circuit.

## Acceptance Criteria

- Irreversible resolves require_human absent a rule; passable only by a sufficiently trusted actor under named constraints.
- Reversibility is still always evaluated and surfaced.
