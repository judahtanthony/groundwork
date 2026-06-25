# Phase 4 — Operator UI (operator-unblock slice)

The minimum server-rendered operator surface for visibility and approvals, built
on the existing `gw server` HTML + SSE dashboard (no SPA; ADR 0042 progressive
complexity). Delivered under `T-1061`:

- **Shell & navigation** (`T-1072`) — shared `layout` template + page pattern;
  activated Tickets and Approvals nav.
- **Tickets / ready / blocked** (`T-1063`) — value-ordered ready queue, blocked
  queue with unmet-dependency annotations, and the full ticket list, reusing the
  CLI's store reads (ADR 0039/0041).
- **Approvals inbox** (`T-1062`) — pending approvals grouped by risk with gate
  reason, requesting actor, required actor/role constraints, and ticket context.
- **Approval decisions** (`T-1064`) — approve / reject / clarify routed through
  the same `ApprovalService` path as the API and CLI (no self-approval, no
  policy bypass; ADR 0028).
- **Land preview** (`T-1073` API, `T-1065` UI) — `GET /api/v1/tickets/{id}/land/preview`
  and an inline staged-diff preview on `land_to_main` approvals.
