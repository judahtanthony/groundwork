# ADR 0031: Server vs Store-Direct Boundary

Status: Accepted

## Context

`cli.md` states that if `gw server` is running, mutating commands should call the
local API by default; if it is not, "simple ticket/config commands may open
SQLite directly through the shared store package," while "commands requiring live
run control must fail clearly without the coordinator." `coordinator.md` repeats
this: SQLite supports multi-process access, but active orchestration must go
through the coordinator to avoid split-brain scheduling, inconsistent approvals,
and duplicate run lifecycle logic. Phase 1 commands always open the store
directly (`internal/cli/store.go`). Phase 2 adds the server, so the boundary —
which commands require it — must be made precise and enforced.

## Decision

Classify every command into **store-safe** and **coordinator-required**, and
route through a small client that prefers the API when the server is reachable.

- **Store-safe (direct SQLite allowed when server is down).** Reads and simple,
  non-orchestration ticket/config mutations: `ticket create/list/show/edit/
  assign/transition/tree/context/link/export/import`, `actor *`, `status`,
  `board`, `context`, `doctor`, `export`. These touch durable operational records
  with the same transactional store and are safe single-process. When the server
  *is* running, they still prefer the API so the running coordinator's caches and
  SSE stream stay coherent.
- **Coordinator-required (must fail clearly without the server).** Anything that
  drives live runs, scheduling, or gate decisions on active work: `run once/next/
  pause/resume/cancel`, `approval approve/reject/clarify`, `validation run`,
  `ticket decompose` and `ticket escalate` (they open planning runs / route
  re-plan decisions), and `server` itself. Without a reachable coordinator these
  return a clear, machine-coded error (e.g. `coordinator_required`) rather than
  mutating the store and risking split-brain.
- **One client, two transports.** A `internal/cli` client tries the HTTP API at
  `config.Server.Addr`; on connection refusal it falls back to the direct store
  **only for store-safe commands**. Coordinator-required commands skip the
  fallback and error. This keeps a single code path per command and makes the
  boundary a property of the command, not of each call site.
- **Read consistency.** Read commands may use either transport; the API response
  and the direct-store response are the same projection, so output is stable
  regardless of whether the server is up.

## Consequences

`internal/cli` gains a coordinator client and a per-command capability flag
(store-safe vs coordinator-required). `gw doctor` reports whether the coordinator
is reachable. The rule operationalizes `coordinator.md`'s split-brain warning:
live run control has exactly one owner. It also preserves the Phase 1 ergonomic
that simple ticket/config work keeps working with no server running, which the
self-hosting phase (M3) relies on.
