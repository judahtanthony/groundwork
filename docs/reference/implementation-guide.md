# Implementation Guide

This is the recommended first implementation sequence after documentation bootstrap.

1. Add Go module and CLI skeleton.
2. Implement config discovery for `.groundwork/config.yaml`.
3. Implement `gw init` to create `.groundwork/` in a managed repo.
4. Add SQLite migrations and store setup.
5. Implement work-node CRUD (uniform nodes with advisory `kind`, operational `work_type`, and structural `node_type`) and export.
6. Implement work-tree records, dependency edges, rollups, and `gw context` briefs.
7. Implement actor registry parsing for `.groundwork/actors.yaml`.
8. Implement transactional claim and lease logic, with dependency-aware eligibility.
9. Implement `gw server` with health and state endpoints.
10. Implement dashboard shell.
11. Implement actor-aware approval records and CLI/web decisions, including the `decompose` gate, reversibility gating, and escalation/re-plan.
12. Implement validation templates, validation results, and canon distillation with parent reconciliation.
13. Implement recovery and stale run handling, resuming runs from their last checkpoint.
14. Import `docs/plan/work-tree.yaml` into Groundwork.
15. Add Codex runtime adapter, planning/implementation runs, isolated worktrees, actor snapshots, and run checkpoints squashed at landing.

Phases 1–2 cover steps 1–8/10–12; decomposition proposals, SOPs, and autonomy elevation build on the approval and policy work. See `docs/product/roadmap.md`.

At each step, update the relevant docs and ADRs if behavior changes.
