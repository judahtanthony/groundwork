# Dashboard

The v1 dashboard is localhost-only, single-user, and server-rendered by Go with minimal JavaScript. It is an operational surface over coordinator state, not a source of truth.

## Navigation

A left sidebar groups pages into **Operate** (Dashboard, Board, Tickets, Runs, Approvals) and **Configure** (Policies, Settings), with a repo/branch header and a server + `state.sqlite` footer. A topbar carries breadcrumbs, search, and a live-status indicator. A narrow/mobile layout collapses to a small bottom tab bar.

## Pages

- Dashboard: KPIs (active runs, blocked, pending approvals, in review, validations failing, landed today), an active-runs table, an attention queue, a recent-event timeline, and resource/runtime panels.
- Board: work nodes grouped by the nine ticket statuses (`backlog` → `done`); drag to transition.
- Tickets: the work-node tree — hierarchy/outline with parent breadcrumbs, leaf/composite indicators, dependencies, and rollup state. Ticket detail covers problem, acceptance criteria, validation requirements, runs, current diff, and timeline.
- Run detail: live transcript, plan, changed files, validations, token/cost metrics, and any linked approval; pause/resume/cancel, open workspace, copy resume command, export transcript.
- Approval inbox: approvals grouped by risk class; approve once, reject, ask the agent to clarify, and approve + suggest rule.
- Policies: trust rules (ordered, with stable IDs), validation templates by file type, a rule editor, and the suggestion queue for policy learning.
- Settings: repository / SQLite paths, agent engine + sandbox mode, concurrency + lease timing, server bind/port, AGENTS.md sync, and `gw doctor` health checks.

## Decomposition Surfaces

The dashboard must expose the work-node model (see [work-tree.md](work-tree.md) and ADRs [0009](../adr/0009-generalized-work-nodes-and-dynamic-decomposition.md)–[0011](../adr/0011-progressive-planning-autonomy-via-sops.md)):

- hierarchy and parent breadcrumbs on the Tickets/Board surfaces, with leaf vs composite indicated;
- a composite node's Design/Contract section and its children with rollup state;
- decomposition proposals as a reviewable approval (`decompose`), gated like landing;
- dependency edges (blocks / blocked-by) and dispatch eligibility;
- escalation / upward-revision entries on a node;
- autonomy levels and task-type SOPs on the Policies surface.

## Risk And Reviewer Display

- Risk shows a 0–100 score mapped onto the named classes (`low`/`medium`/`high`/`critical`); policy gates key off the class, not the raw number. See [trust-and-approvals.md](trust-and-approvals.md).
- Reviewer-agent modes are a Phase 2 surface; v1 exposes only auto and require-human.

## Realtime Updates

Use Server-Sent Events for v1 realtime updates, with a lightweight live/heartbeat indicator. WebSocket is not required unless later interactivity needs it.

## Visual Reference

The v1 visual language and screen layouts come from the design handoff under [../design/](../design/). The work-node surfaces the wireframes do not yet draw are specified in [../design/decomposition-ui-spec.md](../design/decomposition-ui-spec.md).

## Boundary

The dashboard is an operational surface over coordinator state. It is not the source of truth.
