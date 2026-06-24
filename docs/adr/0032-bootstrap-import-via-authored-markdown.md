# ADR 0032: Bootstrap Work-Tree Import via Authored Markdown Exports

Status: Accepted
Implemented: Implemented

## Context

`docs/plan/work-tree.yaml` is the bootstrap planning tree
(`schema: groundwork_bootstrap_work_tree/v1`): goal/epic/ticket nodes with
`G-`/`E-`/`T-` ids, `kind`, `title`, `children`, and `acceptance`. It carries no
`status`, `work_type`, or phase tags. Phase 3 ticket T-1001 (E-0011) requires that
this tree "becomes managed Groundwork tickets" while "imported records preserve
hierarchy."

The M2 cold-start importer (T-0902) rebuilds node rows and dependency edges from
committed Markdown exports under `.groundwork/tickets/**/ticket.md`
(`ticket-export.md`) — it does not read the planning YAML. [ADR 0019](0019-uniform-ticket-ids.md)
already states that the planning ids "do not bind the runtime scheme; they
normalize when that tree is imported (Phase 3)." So Phase 3 must decide (a) the
ingest mechanism, (b) id reconciliation, and (c) how already-completed work
(Phases 0–2) is represented.

## Decision

**No YAML ingest surface.** `work-tree.yaml` stays a planning artifact. The tree is
transcribed once, by hand, into canonical Markdown exports in the
`ticket-export.md` format and ingested through the existing `gw ticket import`.
One ingest contract (Markdown) is maintained, not two.

- **Whole tree, completed work as `done`.** Every `G`/`E`/`T` node is transcribed,
  preserving the parent/child hierarchy. Completed Phase 0–2 nodes
  (E-0001..E-0010 and descendants) carry `status: done`; the Phase 3 dogfooding
  epic E-0011 and the Phase 4–5 nodes carry `backlog`/`todo`. Representing finished
  work as `done` keeps the eligible set clean while preserving the full ancestor
  spine for `gw ticket context`.
- **IDs preserved verbatim.** `G-0001`, `E-00NN`, and `T-NNNN` map unchanged to the
  store's `TEXT PRIMARY KEY`. The export format already permits non-`T-` parent ids
  (the `ticket-export.md` example uses `parent: EPIC-store`), so goal/epic ids need
  no remap. The `T-NNNN` allocator reseeds to `max(T-NNNN) + 1` on import
  ([ADR 0019](0019-uniform-ticket-ids.md)) so newly created tickets cannot collide
  with imported ones.
- **Field mapping.** `kind` (goal/epic/ticket) is preserved as the advisory label;
  `acceptance` arrays become the `## Acceptance Criteria` section; `node_type`
  (leaf/composite) and `work_type` are assigned during transcription from the node's
  role. `depends_on` edges, including cross-subtree edges, are written explicitly.

## Consequences

The runtime never parses the planning YAML; the YAML is throwaway bootstrap input
and future re-planning happens in the managed tree and its exports, not the YAML.
Authoring the exports is a one-time mechanical pass (T-1001). Because the full tree
exercises importer paths the M2 round-trip test did not — `status: done`, goal/epic
non-`T-` parent ids, and cross-subtree dependencies — a hardening/verification
ticket (T-1006) extends the T-0902 round-trip gate (`export` → delete
`state.sqlite` → `import` → identical tree + edges + statuses) to the whole tree.
This keeps "imported records preserve hierarchy" verifiable rather than asserted.
