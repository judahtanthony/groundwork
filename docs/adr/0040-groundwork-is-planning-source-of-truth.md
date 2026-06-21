# ADR 0040: Groundwork Is The Planning Source Of Truth

Status: Accepted

## Context

M3 imported the bootstrap work tree into Groundwork ([ADR 0032](0032-bootstrap-import-via-authored-markdown.md)),
so the live plan now lives in the gw work tree (`.groundwork/tickets/**` exports,
rebuilt into `state.sqlite`). The static planning files — `docs/plan/work-tree.yaml`
and `docs/plan/phase-2-tickets.md` / `phase-3-tickets.md` — are now **redundant** with
that tree and risk confusing future sessions (AGENTS.md "Required Reading" still points
at `work-tree.yaml`). [ADR 0036](0036-work-as-universal-substrate.md) frames planning
itself as work ("update a doc so future planners understand a decision" is a work
node), so the breakdown belongs in the tree, not in parallel static files.

## Decision

**The gw work tree is the source of truth for the plan** — what work exists, its
ordering, status, and acceptance. It is inspected and evolved through `gw`
(`ticket tree`, `create`, `decompose`, `link`, `transition`), not by editing static
breakdown files.

- **Canon vs. plan split.** Product / architecture / contract / ADR docs and
  SOPs/policies remain committed **canon** — the durable diff a completed subtree
  leaves behind ([ADR 0013](0013-canon-as-memory.md)): the *why/what*. gw holds the
  *work to evolve* canon; canon files hold the *result*. Milestones stay as **roadmap
  prose** (narrative altitude), not first-class in gw; phase grouping is derivable from
  the tree + dependencies if ever needed (milestone gate-nodes), and is deferred.
- **Retire the redundant breakdown files** — `work-tree.yaml`, `phase-2-tickets.md`,
  `phase-3-tickets.md`, and `milestones.md` — once their unique content (ADR
  cross-refs, per-ticket gate notes) is ported into gw nodes. **Archive, not delete**
  (move under `docs/plan/archive/` with a pointer): git history plus a breadcrumb, not
  a hole. Keep `roadmap.md` and `vision.md` as canon.
- **Record the operational workflow** in `.groundwork/WORKFLOW.md` (the committed agent
  operating contract): the loop is *orient via the gw tree → pick the next eligible
  node → read its context brief → execute → validate + land → distill into canon +
  record context-misses → re-plan via escalation*. AGENTS.md "Required Reading" points
  at the gw tree, not `work-tree.yaml`.

**Classification ([ADR 0037](0037-transitional-defaults-vs-invariants.md)):** that the
tree is the authoritative plan is structural; the specific file retirement is process.
No invariant changes.

## Consequences

Future sessions orient from AGENTS.md → the gw tree → the eligible set, rather than
reading static breakdown files that drift from reality. The transition is itself tracked
as an initiative in gw (dogfooding the claim). Until the retirement ticket runs, the
historical `phase-N-tickets.md` ids (planning labels, [ADR 0019](0019-uniform-ticket-ids.md))
overlap the runtime allocator's freshly minted ids; the archive step resolves the
overlap, and the runtime tree is authoritative in the meantime.
