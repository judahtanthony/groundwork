# Implementation Guide

This is the recommended first implementation sequence after documentation bootstrap.

1. Add Go module and CLI skeleton.
2. Implement config discovery for `.groundwork/config.yaml`.
3. Implement `gw init` to create `.groundwork/` in a managed repo.
4. Add SQLite migrations and store setup.
5. Implement work-node CRUD (uniform nodes with advisory `kind` and `node_type`) and export.
6. Implement work-tree records, dependency edges, rollups, and `gw context` briefs.
7. Implement transactional claim and lease logic, with dependency-aware eligibility.
8. Implement `gw server` with health and state endpoints.
9. Implement dashboard shell.
10. Implement approval records and CLI/web decisions, including the `decompose` gate, reversibility gating, and escalation/re-plan.
11. Implement validation templates, validation results, and canon distillation with parent reconciliation.
12. Implement recovery and stale run handling, resuming runs from their last checkpoint.
13. Import `docs/plan/work-tree.yaml` into Groundwork.
14. Add Codex runtime adapter, planning/implementation runs, isolated worktrees, and run checkpoints squashed at landing.

Phases 1–2 cover steps 1–8/10–12; decomposition proposals, SOPs, and autonomy elevation build on the approval and policy work. See `docs/product/roadmap.md`.

At each step, update the relevant docs and ADRs if behavior changes.

