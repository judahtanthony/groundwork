---
id: T-1008
kind: ticket
node_type: leaf
work_type: technical_implementation
title: Order the eligible set by value, not FIFO
status: done
assignee: null
requested_actor: null
priority: null
labels: []
parent: T-1006
depends_on:
    - T-1007
created_at: "2026-06-21T20:08:26Z"
updated_at: "2026-06-21T22:35:55Z"
---

## Problem

Per ADR 0039: replace the scheduler ORDER BY id with score(): lexicographic priority path (desc) then id-path (DFS) tiebreak. score() is the seam for the future multi-signal model.

## Acceptance Criteria

- Eligible nodes order by priority path then DFS id-path.
- Unset priority => FIFO/DFS; setting >0 floats the subtree ahead; no descendant writes on re-prioritize.
