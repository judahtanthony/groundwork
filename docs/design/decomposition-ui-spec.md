# Decomposition UI Spec

## Context

The imported [wireframes/](wireframes/) treat tickets as a flat tracker of executable units. They predate the work-node model in ADRs [0009](../adr/0009-generalized-work-nodes-and-dynamic-decomposition.md)–[0011](../adr/0011-progressive-planning-autonomy-via-sops.md) and [work-tree.md](../architecture/work-tree.md), so none of that model has a surface yet. This spec defines what the web app must add to close that gap. It is the source of truth where it differs from the prototypes; the prototypes remain the reference for layout, spacing, and visual language.

The model in one paragraph: work is a uniform tree of **nodes** (`kind` is an advisory label; `work_type` drives SOPs/policy/routing; the structural fact is **leaf** vs **composite**, decided by a **triage** at claim time). Composite nodes are decomposed just-in-time into children plus a **parent contract**; decomposition is a reviewable **proposal** gated like landing. **Dependency edges** form a DAG overlay that governs dispatch eligibility. **Escalation** propagates revisions upward. `decompose`, `execute`, and `land_to_main` are gated actions whose **autonomy level** loosens as work-type **SOPs**, context, and validations mature.

## New shared primitives

Add to the component inventory ([wireframes/screen-components.jsx](wireframes/screen-components.jsx)):

- **Node-type indicator** — leaf vs composite (composite implies it has/expects children + a contract).
- **Kind chip** — advisory label (`goal`/`epic`/`ticket`/`task`/…); visually subordinate to the node-type indicator.
- **Dependency badge** — `blocked-by N` / `blocks N`, plus an "eligible / waiting on deps" state distinct from the `blocked` status badge.
- **Escalation marker** — an upward-revision flag on a node and its timeline.
- **Autonomy-level badge** — per gated action (`execute` / `decompose` / `land`): human-required → reviewer (Phase 2) → auto.
- **Actor badge** — human, AI agent, AI judge, or tool, with actor id, role/capability summary, and runtime/model where relevant.
- **Decompose approval card** — an approval variant whose body is a proposed plan (children + contract + edges), not a command/diff.

## Per-screen gaps

### Tickets — new work-tree view
The `Tickets` nav item currently has only a detail screen. Add a tree/outline view:
- expandable hierarchy with parent **breadcrumbs**; leaf vs composite indicated per row;
- **rollup** state on parent rows (derived from children per [work-tree.md](../architecture/work-tree.md));
- dependency indicators and an "eligible to dispatch" vs "waiting on deps" distinction;
- filters by kind, status, actor, work type, and "leaf only" (the primary unit operators act on).

### Board ([wireframes/screen-board.jsx](wireframes/screen-board.jsx))
The nine status columns are correct. Add to cards:
- leaf/composite indicator and dependency badge (`blocked-by` should read distinctly from the `blocked` column);
- optional grouping by parent; a composite card should link to its children rather than implying it is directly executable.

### Ticket detail ([wireframes/screen-ticket.jsx](wireframes/screen-ticket.jsx))
- header: node-type + kind; a **triage** banner when a claimed node is being classified;
- **composite nodes**: a **Design / Contract** panel (schemas/interfaces/requirements children depend on) and a **Children** panel with per-child status + rollup; a **Decompose** action;
- **Dependencies** panel: `depends_on` (blocks / blocked-by) with eligibility;
- **Escalations** panel: upward-revision entries + re-plan decisions (human-gated in v1); an **Escalate** action in the action rail;
- right-rail fields gain `kind`, `work_type`, `node_type`, `requested_actor`, and `depends_on`;
- actor routing: show the requested actor, eligible actors, and any policy reason preventing a requested actor from claiming the node.

### Run detail ([wireframes/screen-run.jsx](wireframes/screen-run.jsx))
- a **run-mode** badge: **planning (decomposition)** vs **implementation** (per [runtime-model.md](../architecture/runtime-model.md));
- an **actor summary**: actor id, actor type, runtime/model, and link to the actor snapshot captured for the run;
- for a planning run, the "Changed files / diff" region is replaced by a **proposed plan** preview (children + contract + edges) headed for `review`.

### Approvals inbox ([wireframes/screen-approvals.jsx](wireframes/screen-approvals.jsx))
- add **`decompose`** as an approval type: the detail pane shows the proposed children, parent contract, and dependency edges, with approve / request-changes / reject (children become dispatchable only on approval);
- surface **escalation / re-plan** decisions in the same inbox;
- show requesting actor, required actor/role constraints, and deciding actor on approval records;
- show the relevant **autonomy level** as context on each item (why this still needs a human).

### Policies ([wireframes/screen-policies.jsx](wireframes/screen-policies.jsx))
- add **`decompose`** to the gated actions alongside `execute` and `land_to_main`;
- represent the **autonomy ladder** per action (human-required → reviewer → auto), where reviewer is visibly **Phase 2 / disabled in v1**;
- add an **SOPs** surface: work-type SOPs and context under `.groundwork/sops/<work-type>/`, with the maturity signals (SOP + context + validations) that justify elevating an action's autonomy. Elevation remains an explicit human act; learning may only *suggest* it (reuse the existing suggestion queue).
- add an **Actors** surface for `.groundwork/actors.yaml`: list actors, create/edit human and AI actors, validate actor ids/types/roles/capabilities, configure AI runtime/model/sandbox, and preview which policies each actor can satisfy.

### Settings ([wireframes/screen-settings.jsx](wireframes/screen-settings.jsx))
- add the `.groundwork/sops/` path alongside the existing repository/SQLite paths.
- add the `.groundwork/actors.yaml` path and a link to the actor creation/registry surface.

## Definition of done

The web surface honors this spec when: a composite node can be triaged, decomposed, and its proposal reviewed; children dispatch only after the proposal is approved and their dependencies are satisfied; a node can escalate upward and a human can re-plan; and `decompose` appears as a gated action with an autonomy level and SOP-backed loosening, consistent with [trust-and-approvals.md](../architecture/trust-and-approvals.md).
