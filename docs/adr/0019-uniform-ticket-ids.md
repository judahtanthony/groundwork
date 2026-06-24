# ADR 0019: Uniform Monotonic Ticket IDs (`T-NNNN`)

Status: Accepted
Implemented: Partial

> Amended by ADR 0032: the one-time bootstrap import preserved historical
> `G-`/`E-`/`T-` IDs verbatim. The uniform `T-NNNN` allocator remains the rule
> for newly created nodes.

## Context

Work is a uniform tree of nodes and `kind` is advisory (ADR 0009). Node IDs must be human-typeable on the CLI (`gw ticket show T-0123`), stable across export and reimport, and collision-free for a single-user local coordinator (ADR 0005). The retired bootstrap planning document (`docs/plan/work-tree.yaml` in git history) used kind-encoded prefixes (`G-`/`E-`/`T-`), but that scheme predates the uniform-node decision.

## Decision

Newly created work nodes get a single uniform prefix `T-` followed by a zero-padded, monotonically increasing integer (`T-0001`, `T-0002`, …), allocated transactionally in SQLite. The advisory `kind` is **not** encoded in new IDs. Exported front-matter `id` is authoritative; on import the allocator is seeded to `max(existing T-NNNN) + 1`. The numeric width grows past four digits on exhaustion.

## Consequences

The counter lives in SQLite and must be reseeded on import. Allocation is deterministic and safe under the single-writer coordinator assumption (ADR 0005). Human-readable IDs are preferred over ULID/UUID for CLI ergonomics. Historical `G-`/`E-` IDs from the bootstrap import remain valid imported IDs per ADR 0032, but the allocator does not mint new non-`T-` IDs.
