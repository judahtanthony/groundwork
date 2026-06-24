# ADR 0025: HTTP Server And SSE Transport

Status: Accepted
Implemented: Partial

## Context

Phase 2 introduces `gw server`, the localhost coordinator, which must expose the
HTTP API in [http-api.md](../contracts/http-api.md) and a Server-Sent Events
stream for the (later) dashboard. The API needs path parameters
(`/api/v1/tickets/{id}`, `/api/v1/runs/{id}/events`) and method-aware routing.
`conventions.md` prefers the Go standard library when clean enough and otherwise
small, mature, open-source dependencies, and forbids mandatory external services
in v1 ([ADR 0008](0008-prefer-open-source-local-dependencies.md)).

## Decision

Use the **Go standard library `net/http` with the 1.22+ `ServeMux`** for routing.
The enhanced `ServeMux` supports method + path patterns and `{id}` wildcards
(`mux.HandleFunc("GET /api/v1/tickets/{id}", ...)`, read via
`r.PathValue("id")`), which covers the entire contract with no third-party router.
The module already targets Go 1.26, so this is available.

- **One server, one bind.** Bind `config.Server.Addr` (default `127.0.0.1:4500`,
  already in `internal/config`). Localhost-only, single-user ([ADR 0005](0005-localhost-single-user-v1.md)); no TLS, no auth in v1.
- **Error envelope is reused.** Responses use the existing
  `{"error":{"code","message"}}` shape (already implemented in
  `internal/cli/app.go`); the two surfaces share one envelope.
- **SSE over `GET /api/v1/events`** uses plain `http.Flusher` — `Content-Type:
  text/event-stream`, `id:`/`event:`/`data:` frames, periodic heartbeat comments
  to keep the connection live, and `id:` frames compatible with a future replay
  buffer. `Last-Event-ID` is accepted in v1 but not replayed because the hub keeps
  no event history.
  WebSocket is not used (`dashboard.md`). The event source is the in-process hub
  defined in [ADR 0026](0026-coordinator-concurrency-model.md).
- **Handlers are thin.** HTTP handlers decode the request, call a
  coordinator/store service, and encode the result or envelope. Business logic
  (scheduling, gate decisions, distillation) lives in services, not handlers
  (the boundary rule in `overview.md`). Projections that are also exported reuse
  `internal/encoding` for deterministic JSON.

A new `internal/server` package holds the mux, handlers, and the SSE writer. The
server-vs-store-direct boundary (which CLI commands require a running server) is
[ADR 0031](0031-server-vs-store-direct-boundary.md).

## Consequences

`internal/server` is added to the architecture map. No new runtime dependency is
introduced. The SSE contract (heartbeat + monotonic event ids) lets the dashboard
maintain a live connection today and leaves a replay-buffer seam for later
`Last-Event-ID` resume. The server-rendered dashboard shell shipped after M2 as
the interim operational UI. Policy-learning suggestion endpoints
(`/api/v1/policies/suggestions/*`) remain future work, consistent with the roadmap
deferring policy learning.
