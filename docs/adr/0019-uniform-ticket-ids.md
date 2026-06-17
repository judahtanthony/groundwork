# ADR 0019: Uniform Monotonic Ticket IDs (`T-NNNN`)

Status: Accepted

## Context

Work is a uniform tree of nodes and `kind` is advisory (ADR 0009). Node IDs must be human-typeable on the CLI (`gw ticket show T-0123`), stable across export and reimport, and collision-free for a single-user local coordinator (ADR 0005). The bootstrap planning document (`docs/plan/work-tree.yaml`) uses kind-encoded prefixes (`G-`/`E-`/`T-`), but that scheme predates the uniform-node decision.

## Decision

Every work node gets a single uniform prefix `T-` followed by a zero-padded, monotonically increasing integer (`T-0001`, `T-0002`, …), allocated transactionally in SQLite. The advisory `kind` is **not** encoded in the ID. Exported front-matter `id` is authoritative; on import the allocator is seeded to `max(existing) + 1`. The numeric width grows past four digits on exhaustion.

## Consequences

The counter lives in SQLite and must be reseeded on import. Allocation is deterministic and safe under the single-writer coordinator assumption (ADR 0005). Human-readable IDs are preferred over ULID/UUID for CLI ergonomics. The `G-`/`E-` IDs in `work-tree.yaml` remain pre-implementation planning labels that do not bind the runtime scheme; they normalize when that tree is imported (Phase 3).
