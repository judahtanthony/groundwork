# ADR 0017: Use The Pure-Go `modernc.org/sqlite` Driver

Status: Accepted
Implemented: Implemented

## Context

Groundwork needs a `database/sql` SQLite driver (ADR 0002). The two practical options are the cgo `mattn/go-sqlite3` (fast, battle-tested, requires a C toolchain) and the pure-Go `modernc.org/sqlite` (no cgo). ADR 0001 wants a small portable binary; ADR 0008 prefers open-source local dependencies; the runtime is single-user and localhost (ADR 0005), so raw throughput is not a constraint.

## Decision

Use `modernc.org/sqlite` and build with `CGO_ENABLED=0`. Apply the pragmas from `docs/contracts/sqlite-schema.md` — `journal_mode=WAL`, `foreign_keys=ON`, `busy_timeout=5000` — on every connection via DSN/PRAGMA.

## Consequences

Static, trivially cross-compiled binaries and no C toolchain in the build or CI. Throughput is lower than the cgo driver but irrelevant at single-user scale. The `database/sql` boundary keeps the driver swappable if performance ever becomes a concern.
