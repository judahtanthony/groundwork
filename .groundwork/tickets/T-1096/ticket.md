---
id: T-1096
kind: ticket
node_type: null
work_type: technical_design
title: Prune/archive settled work to reduce board noise
status: backlog
assignee: null
requested_actor: null
priority: null
labels: []
parent: T-1074
depends_on: []
created_at: "2026-07-05T18:37:38Z"
updated_at: "2026-07-05T18:37:38Z"
---

## Problem

Completed tickets accumulate and add noise to the TUI/GUI/CLI. Recommendation (phased):

1. NEAR-TERM (no schema change): default views (gw ticket list/tree/board, TUI, GUI) hide nodes whose entire subtree is settled (done/cancelled) under a settled root; a --all / 'show archived' toggle reveals them. Extends the spirit of T-1035. Highest value, cheapest.
2. MEDIUM: add an 'archived' terminal status distinct from done/cancelled, reached via 'gw ticket archive <id>' (+ unarchive). Excluded from default views, parent rollups, and eligibility. Archiving a root archives its settled subtree. File-authoritative like every other status (ADR 0053).
3. LATER: harvest-before-archive. When a root/subtree completes, the system suggests a distillation task (review/documentation work type) that harvests durable learnings into canon (ADRs/docs/SOPs) and the parent context before archiving (ADR 0013/0035/0047). Ties into the review-agent and progressive-planning futures.

Never delete: completed tickets are durable canon evidence that rebuilds the store. Pruning = hide/archive, not delete. Related: T-1035 (reconcile completed composites), T-1020 (store rebuild on cold-start reads).

## Acceptance Criteria

- Default TUI/GUI/CLI views hide fully-settled subtrees, with a reveal toggle (--all)
- An explicit archived terminal state exists (gw ticket archive/unarchive), excluded from default views, rollups, and eligibility
- Settled work is never deleted; archiving preserves the durable sidecars/decisions/summaries
- A harvest-before-archive step distills learnings into canon before a completed subtree is archived
