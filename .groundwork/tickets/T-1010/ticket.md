---
id: T-1010
kind: ticket
node_type: leaf
work_type: technical_implementation
title: 'Reversibility: highest-bar condition, not pre-policy short-circuit'
status: done
assignee: human.owner
requested_actor: null
priority: 0.9
labels: []
parent: T-1009
depends_on:
    - T-1067
created_at: "2026-06-21T20:08:43Z"
updated_at: "2026-06-26T00:46:14Z"
---

## Problem

Per ADR 0038: gate.go evaluates reversibility through rules as a very-high-bar condition (still always computed/surfaced, an ADR 0037 invariant), not a return-before-policy short-circuit.

## Acceptance Criteria

- Irreversible resolves require_human absent a rule; passable only by a sufficiently trusted actor under named constraints.
- Reversibility is still always evaluated and surfaced.
