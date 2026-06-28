# ADR 0053: Filesystem-Authoritative Durable State

Status: Accepted
Implemented: Partial

## Context

Groundwork's long-lived project state must survive deletion of
`.groundwork/state.sqlite`. ADR 0012 already provides the right test: if losing a
datum would strand work or change project meaning, it must be durable in files.
However, several docs and code paths still treat SQLite as primary during runtime
and ticket exports as a later projection. For example, `gw ticket create` and the
HTTP ticket create endpoint currently insert only into SQLite; the new ticket is
lost on a database purge unless a separate export runs.

That is a source-of-truth mismatch. SQLite is valuable for transactional claims,
graph queries, leases, and scheduler coordination, but it should not be the only
home for tickets, blockers, approvals that must survive rebuild, SOPs, policies,
actors, workflow, or canon.

## Decision

Durable project state is filesystem-authoritative. SQLite is a transactional
projection/cache over durable files plus a store for ephemeral runtime state.

Filesystem-authoritative durable state includes:

- ticket records and dependency edges,
- ticket-attached decision/input/approval/rework/recovery records,
- SOPs and work-type context,
- trust, validation, risk, and autonomy policies,
- actor registry,
- workflow instructions,
- canon: ADRs, architecture docs, contracts, product docs, and code,
- committed landing outputs.

SQLite may be authoritative only for runtime state whose loss is safe or
recoverable from durable files and git:

- active leases,
- process ids,
- live queue handles,
- scheduler bookkeeping,
- transient run status for currently executing processes,
- ignored run logs, transcripts, and raw artifacts,
- generated views.

Mutations to durable state must not report durable success until the filesystem
source of truth has been updated, or until an explicit durable write-ahead record
has been written that can complete/replay the export after a crash. Async export
is allowed only behind that durability boundary; "queued in SQLite" is not enough.

On startup, Groundwork rebuilds SQLite from durable files and git. If SQLite and
durable files disagree, recovery must detect and surface the divergence rather
than silently treating SQLite as newer truth. The normal repair path is to
rebuild SQLite from files; any unexported SQLite-only durable mutation is a bug
or a recovery-needed condition.

## Consequences

This ADR supersedes the "SQLite-primary, exported to files" phrasing in older
docs for durable operational records. SQLite remains the live coordination
database, but durable state is file-authoritative and SQLite-projected.

Implementation must add write-through or durable-journaled export behavior for
ticket mutations such as create, edit, transition, link/unlink dependencies,
reparent, decompose, escalate, decision/input/approval records, and rework or
recovery-state changes. Tests should prove that after each durable mutation,
deleting `.groundwork/state.sqlite` and rebuilding from files preserves the
meaningful state.
