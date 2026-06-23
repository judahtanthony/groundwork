---
id: T-1032
kind: ticket
node_type: leaf
work_type: technical_implementation
title: Make import->export byte-stable for the empty-timestamp convention
status: done
assignee: null
requested_actor: null
priority: 0.65
labels:
    - cli-ux
    - determinism
parent: T-1022
depends_on: []
created_at: "2026-06-22T19:28:34Z"
updated_at: "2026-06-23T20:19:39Z"
---

## Problem

Diagnosis refined: import->export is byte-stable in code — ImportTicket preserves timestamps verbatim (incl. the empty-string convention, ADR 0021) and DependencyIDs returns deps ORDER BY to_id. The early 50-file churn was a stale pre-existing store. The one real residual: T-1003's committed export was authored before dep-ordering determinism, so its depends_on order ([T-1002,T-0503,T-0504]) churns to id-sorted on rebuild. Fix: normalize the stale export and add a cold-rebuild round-trip regression test locking byte-stability.

## Acceptance Criteria

- A cold rebuild (import from committed exports) re-exports byte-identically; covered by a round-trip regression test exercising empty timestamps and out-of-order deps
- The stale committed export (T-1003) is normalized to the deterministic dep order
