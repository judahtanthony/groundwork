# ADR 0018: Embedded, Ordered, Forward-Only SQL Migrations

Status: Accepted

## Context

`docs/contracts/sqlite-schema.md` requires migrations that apply in order and are safe to re-run. Dedicated migration libraries (`golang-migrate`, `goose`) add a dependency and features Groundwork does not need.

## Decision

Embed numbered SQL files (`internal/store/sqlite/migrations/0001_init.sql`, …) with `embed.FS`. A small runner applies pending files in lexical order, each in its own transaction, and records `id`, `applied_at`, and a `checksum` in a `schema_migrations` table. Re-running is a no-op; a changed checksum on an already-applied migration is a hard error (drift detection). Migrations are forward-only in v1.

## Consequences

No migration dependency, deterministic behavior, and full ownership of the runner. Forward-only is acceptable because the development database is disposable (ADR 0007, ADR 0012): rolling back during development means deleting `state.sqlite`. Down-migrations can be added later if a concrete need appears.
