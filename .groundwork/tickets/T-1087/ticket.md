---
id: T-1087
kind: ticket
node_type: null
work_type: technical_implementation
title: Harden startup divergence check and orphaned in_progress recovery
status: backlog
assignee: null
requested_actor: null
priority: null
labels: []
parent: T-1074
depends_on: []
created_at: "2026-07-05T16:53:11Z"
updated_at: "2026-07-05T16:53:11Z"
---

## Problem

Two pre-existing startup rough edges surfaced by the Phase 6 Codex demo (spurious 'diverge from their sidecars' warning on a stuck T-0002).

1. Boot ordering: internal/cli/server.go runs db.DetectFileDivergence() BEFORE db.ReconcileStartup(). A node transiently in_progress (a live or interrupted run) is about to be requeued to todo by reconcile, but the earlier divergence check already flags SQLite(in_progress) != sidecar(todo) as recovery_needed. Move DetectFileDivergence to run AFTER ReconcileStartup so requeued nodes match their sidecars first.

2. Orphaned in_progress: ReconcileStartup only requeues in_progress nodes that still hold a lease. An interrupted/failed run whose lease was released leaves the node stuck in_progress forever (txClaim sets in_progress without write-through, so its committed sidecar still says todo -> permanent divergence). At startup no run is live, so reconcile should requeue ANY in_progress node to todo, healing the stuck case. See internal/store/sqlite lease.go (txClaim) + the ReconcileStartup/DetectFileDivergence code.

Deferred from Phase 6 (PR #1); not a blocker for merge.

## Acceptance Criteria

- Restarting gw server while a node is mid-run (or after an interrupted run) does not emit a false 'N ticket(s) diverge' warning
- A node left stuck in_progress by an interrupted/failed run is requeued to todo at boot and lands cleanly
- Regression tests cover both boot-ordering and orphaned-in_progress recovery
