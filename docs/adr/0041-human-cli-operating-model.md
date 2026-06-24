# ADR 0041: Human-CLI Operating Model And Command Surface

Status: Accepted
Implemented: Implemented

## Context

[ADR 0040](0040-groundwork-is-planning-source-of-truth.md) made the gw work tree the
planning source of truth, and [ADR 0039](0039-value-prioritization-v1.md) gave the
scheduler a value signal so the eligible set runs in priority order. But the human who
drives that loop today — before the web dashboard ([E-0009](../../.groundwork/tickets/E-0009/ticket.md))
or the AI runtime ([E-0006](../../.groundwork/tickets/E-0006/ticket.md)) exists — has
no read surface onto that engine. Dogfooding surfaced the gap concretely:

- **Eligibility is computed but unreadable.** `sqlite.ListEligible` (todo + dependencies
  satisfied) and the scheduler's value ordering (`orderByValue`/`priorityPath` in
  `internal/scheduler/priority.go`) already answer "what is ready, in priority order,"
  but that logic is **locked inside the `Scheduler`** and reachable only by machine
  dispatch (`gw run next`, policy-gated). `gw ticket list --status todo` ignores
  dependencies entirely — it lists `T-0503` (blocked by the todo `T-0502`) as if it were
  workable. A human cannot tell ready from blocked without hand-tracing the DAG.
- **No "what should I pick up?" for a person.** `gw run next` is machine dispatch to AI
  actors, not a human picker.
- **Help is structurally terse.** `internal/cli/command.go` hard-codes every leaf
  command's help to `Usage:\n  <cmd> [--json]`; the `Command` type has no field to
  declare flags, so `list`/`create`/`edit`/`board` cannot advertise `--status`,
  `--priority`, `--parent`, … .
- **`--json` is inconsistent.** `gw ticket tree --json` omits `priority` (and `parent`,
  `work_type`) that `gw ticket show --json` includes.
- **No guided assignment, reparenting, or landing preview.** `assignee` exists but is a
  display-only label set through a generic `edit`; parent is fixed at creation; the
  human approves a landing without a built-in way to preview the diff.

This ADR records the **human-CLI operating model** — the commands a person uses to drive
the orient → pick → claim → execute → land loop ([WORKFLOW.md](../../.groundwork/WORKFLOW.md))
— and the principles the new surface must hold to.

## Decision

**The CLI exposes the existing eligibility/priority engine as first-class human reads and
guided transitions, computed from one shared ordering, without loosening any gate.**

### Principle: one ordering, two consumers

The value ordering (`orderByValue`/`priorityPath`/`comparePath`) moves out of
`internal/scheduler` into a shared, store-backed surface (e.g.
`DB.ListEligibleOrdered()`): "todo + dependencies satisfied, ordered by the ADR 0039
priority path." Both the scheduler and the CLI read path consume it, so the human "what's
ready" answer and the machine dispatch order are the **same computation**, not a CLI
re-implementation that can drift from ADR 0039.

### Read surface — "what's the state, what can I do, what needs me"

- **`gw next`** — the single top eligible node (value-ordered) plus a compact context
  brief (ancestor spine, acceptance, dependencies) and the command to take it. The
  honest answer to "what should I pick up?" `--claim` picks and claims the top node in
  one step. `--json` parity.
- **`gw ticket list --ready`** — the full eligible set, value-ordered.
  **`gw ticket list --blocked`** — todo nodes with unsatisfied dependencies, annotated
  with which dependencies are not yet done. These fold into the existing `list` rather
  than adding top-level verbs.
- **`gw status`** gains eligible (ready), blocked, and pending-approval counts, so one
  command answers "what's the state?", "what can I do?", and "what needs me?".

### Assignment and execution — guided transitions

- **`gw ticket claim <id> [--actor <id>]`** — verifies the node is eligible, sets
  `assignee`, transitions `todo → in_progress`, and prints the brief and next-step hint
  in one guided step. It refuses ineligible/blocked nodes with a clear message. Claiming
  is convenience over the existing primitives (`edit --assignee` + `transition`), not a
  new authority.
- **Reparenting** is added to the existing edit as **`gw ticket edit --parent <id>`**:
  parent-existence check, cycle prevention (a node may not be parented under its own
  descendant), recomputed rollups on the old and new parents, and a `reparented` audit
  event — never a hand-edit of the export.

### Review and landing

- **`gw ticket land <id> --preview`** (dry-run) shows the staged change set the
  `land_to_main` gate would commit, without opening the approval — so the human reviews
  the diff before deciding. The approve/reject flow and the human-gated landing are
  unchanged.

### Consistency

- The `Command` type gains a way for leaf commands to declare their flags/usage, and
  `printHelp` renders them instead of the hard-coded `[--json]`. Leaf commands backfill
  real flag lists, making `-h` self-documenting.
- `gw ticket tree --json` includes `priority` (and `parent`, `work_type`) to match
  `show --json`.

### What this is *not*

The picker recommends; it never claims or lands on its own. Surfacing the eligible set,
the next pick, pending approvals, and the landing diff is **visibility, not authority** —
no command added here self-approves, auto-claims without `--claim`, or loosens a gate.
Landing stays human-approved ([ADR 0034](0034-minimal-git-landing.md)); priority orders
but never authorizes ([ADR 0039](0039-value-prioritization-v1.md)); default-deny
([ADR 0028](0028-gate-evaluation-engine.md)) is untouched. The surface is **domain-agnostic**
([ADR 0036](0036-work-as-universal-substrate.md)): it reads uniform work nodes (ready,
blocked, next), not domain concepts.

**Classification ([ADR 0037](0037-transitional-defaults-vs-invariants.md)):** the command
surface and ergonomics are **process** — they make the human loop legible and may be
reshaped as the dashboard and runtime arrive. That the reads are visibility-only and the
landing gate stays human-approved is the **invariant** they must not cross.

## Consequences

- The eligibility engine that has existed since [ADR 0024](0024-dependency-satisfaction-and-rollup-terminality.md)
  finally has a human surface; `gw next` / `gw ticket list --ready` replace hand-tracing
  the DAG, and `--blocked` makes "blocked by what?" answerable at a glance.
- Extracting the ordering to a shared surface removes the scheduler's private copy as the
  only home for ADR 0039 ordering; the scheduler is refactored to consume it
  (behavior-preserving, covered by existing tests).
- `gw run next` (machine dispatch) and `gw next` (human picker) coexist with distinct
  audiences; the overlap in name is intentional — same question, different actor.
- This is the first substantial *code* surface taken through Groundwork's own gate
  (toward milestone [T-1003](../../.groundwork/tickets/T-1003/ticket.md)); the epic's ADR
  ticket lands first, then the foundational ordering extraction, then the reads and
  guided verbs.
- It builds on [ADR 0016](0016-cli-stdlib-flag-router.md) (the stdlib-`flag` subcommand
  router) by giving leaf commands a way to declare their own flags, and it gives
  [ADR 0039](0039-value-prioritization-v1.md) the human-facing surface it deferred. The
  guided verbs keep claiming/landing as convenience over policy, consistent with
  progressive autonomy via SOPs ([ADR 0011](0011-progressive-planning-autonomy-via-sops.md)):
  ergonomics never become authority.
