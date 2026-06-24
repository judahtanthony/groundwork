# ADR 0042: Web UI Architecture — Contract-First API, Embedded Static SPA, Progressive Realtime

Status: Accepted

## Context

[dashboard.md](../architecture/dashboard.md) established a localhost-only,
server-rendered v1 dashboard, and [T-0801](../../.groundwork/tickets/T-0801/ticket.md)
shipped its shell (server-rendered Go HTML, full-page reload on SSE events). The web
UI is expected to become the **primary human interface** for Groundwork and to grow
well past an operational dashboard: full parity with the CLI ([ADR 0041](0041-human-cli-operating-model.md)),
genuinely realtime-dynamic updates, and — as the product scales — multiple teams, many
concurrent agents, remote agent runners, and an **embedded coordinator/admin agent** (a
conversational, streaming,
tool-using surface inside the app).

Those requirements (token-level streaming, agent chat, optimistic editors, high-frequency
multi-agent realtime) sit above the ceiling of a server-rendered / hypermedia approach:
past a point, hand-rolling that interactivity is rebuilding a client framework poorly.
But a naive jump to a Single-Page App threatens two properties we must keep: the server
as the single source of truth ([ADR 0031](0031-server-vs-store-direct-boundary.md)), and
the **single-binary, local-first distribution** — today a user installs one `gw` binary,
runs `gw server` in a managed repo, and has the whole system from one executable.

## Decision

**The durable commitment is a UI-agnostic contract, not a framework; the client grows in
complexity only as requirements demand, and is always shipped embedded in the `gw` binary.**

1. **Contract-first.** The coordinator's JSON API plus a realtime **event stream** is the
   UI contract (it already largely exists: [ADR 0025](0025-http-server-and-sse-transport.md)).
   Transport is **SSE for v1** and **WebSocket-capable** later where bidirectional or
   high-frequency interaction demands it (agent chat, many concurrent runs). The client
   technology is downstream of and replaceable behind this contract.

2. **Progressive complexity.** Keep the current server-rendered Go HTML for the existing
   operational dashboard — it is cheap and already built (the documented **interim**, not
   throwaway). Adopt a client framework only when a surface actually needs it (realtime
   streaming, the agent-chat surface, rich editors, optimistic UI). Simplicity holds until
   the requirements break it.

3. **Target client: a static-built SPA**, lean (e.g. React or Svelte + Vite), consuming the
   API + event stream — **not** a runtime-SSR meta-framework (Next/SvelteKit-with-a-node-server),
   because a Node runtime beside `gw` would break the single-binary model. State authority
   stays server-side; the client cache is ephemeral and invalidated on stream events (the
   "source of truth" boundary is a discipline, upheld regardless of framework).

4. **Distribution invariant — one binary, no runtime Node.** The SPA is compiled to static
   assets at **build/CI time** and embedded into `gw` via `go:embed`; `gw server` serves
   them same-origin alongside the API and stream. **Users install one binary and run it
   locally — no Node, no package manager, no build step at runtime**, exactly as today
   (the npm/Vite toolchain is a developer/CI concern only). Self-host fonts/assets so there
   is no network dependency. Release prebuilt binaries; contributors build the frontend via
   a `make` target (or a committed `dist/`). This scales unchanged to remote/multi-team: the
   same binary binds beyond localhost (with auth added then).

5. **Binary-size guardrail.** Because the SPA rides inside the binary, the build warns when
   `gw` exceeds **100 MB**, so embedded-asset or dependency inflation is caught early.

6. **Realtime rendering progression.** Full-page reload (current) → in-place region/fragment
   updates driven by the event stream, landing as the SPA does. No surface regresses below
   "data is live."

7. **The web UI is a client of the coordinator, never a bypass.** Every mutation flows
   through the same gates as the CLI — landing and decomposition stay human-gated, approvals
   are never self-granted ([ADR 0028](0028-gate-evaluation-engine.md)/[0034](0034-minimal-git-landing.md)).

**Classification ([ADR 0037](0037-transitional-defaults-vs-invariants.md)):** the API/event
contract, the single-binary local-first distribution, and gate-routing are near-structural;
the specific client framework and the SSE→WebSocket transport step are **process** —
replaceable behind the contract.

## Consequences

- Extends [dashboard.md](../architecture/dashboard.md); the screen map there (Dashboard,
  Board, Tickets, Run detail, Approvals, Policies, Settings) becomes the CLI-parity target,
  built on the SPA as web-UI work activates.
- Sequenced **below** the Codex runtime (E-0006): the runtime is the active focus; this is
  the tracked plan that follows. The current dashboard remains the interim surface meanwhile.
- The single-binary install/run story is **unchanged** for users; only the build pipeline
  gains a frontend step.
- The embedded coordinator/admin agent surface (speculative) has a natural home in a
  streaming-capable SPA over a WebSocket-upgraded stream.
- Recorded as the web-UI epic in gw; the binary-size guardrail ships first as a standalone,
  immediately-useful build check.
