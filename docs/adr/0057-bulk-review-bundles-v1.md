# ADR 0057: Bulk Review Bundles v1

Status: Accepted
Implemented: Partial

Refines [ADR 0047](0047-hierarchical-memory-and-context-compression.md) (completion
summaries, parent memory) and [ADR 0045](0045-node-branching-and-parent-integration.md)
(parent integration) into the minimal v1 review surface for Phase 5. Pairs with envelopes
(ADR 0054) as the "evidence out" boundary opposite the "approval in" boundary.

## Context

Envelopes let a human approve a bounded batch of child work up front (ADR 0054/0056). The
complement is review: instead of inspecting every child, the human should review the
**feature once** at the parent/root boundary, with the evidence compiled for them. Phase 5
must provide this in **manual mode**, before the Phase 6 runtime produces transcripts and
result refs.

## Decision

Add a **review bundle**: a deterministic, read-only assembly over a composite/root
subtree, presented for the human's feature-level decision (the root `land_to_main`
approval, which stays human-gated in v1 and merges the root integration branch to `main`
per [ADR 0058](0058-integration-targets-and-landing-levels-v1.md)).

### Bundle contents

```yaml
review_bundle:
  node_id: T-2000
  envelope: env-T-2000          # the approved boundary this work ran inside
  children:
    - node_id: T-2001
      completion_summary: …      # outcome, decisions, assumptions, risks (ADR 0047)
      changed_files: […]
      diff_ref: …                # staged/landed diff for the child (ADR 0045 result)
      validation: [{command, status}]
      reviewer_findings: […]     # advisory, from the reviewer role (ADR 0055)
      exceptions: […]            # envelope exceptions raised/resolved
  aggregate_diff_stat: …
  unresolved_exceptions: […]
  scope_check: {planned, actual, within: true|false}   # actual arrives with the runtime (ADR 0046)
  recommendation: land | hold | rework
```

### v1 sources (pre-runtime)

- **Completion summaries** become a required output for any child that reaches
  review/landing (ADR 0047). In manual mode the human/coding role records a compact
  summary per child (or it is derived from the landed commit message + validation
  records). The bundle is then a deterministic projection over the subtree's existing
  records — children, dependency edges, validation results, approvals/exceptions — not a
  new data source.
- **Diffs**: per-child landed/staged diff via the existing landing/preview path
  (ADR 0034, the Phase 4 land-preview surface).
- **Reviewer findings**: recorded by the reviewer role (ADR 0055), advisory only.
- **Scope check**: planned scope now; actual-vs-planned comparison fills in with the
  runtime diff (ADR 0046).

### Surfaces

- CLI: `gw review bundle <node>` (`--json`), reusing context-assembly machinery.
- Web: a parent/root review screen built on the Phase 4 server-rendered operator UI
  (sidebar entry, SSE refresh), grouping children with expandable summary → diff →
  validation, and exceptions grouped under the envelope.

### Decision boundary

The bundle **informs** the human's root `land_to_main` decision; it does not auto-approve
anything in v1. Reviewer findings, validation pass/fail, exception count, and scope-check
status are shown so the human can decide once for the whole feature. Auto-approval of
low-risk in-scope child landings, sampling, and reviewer-agent approval are explicitly
**deferred** to later ADRs once track-record data exists (ADR 0043 governance; roadmap
Phase 7+ earned autonomy).

## Consequences

- Completion summaries become a required artifact at child review/landing; the bundle is a
  deterministic read over subtree records (no new authority).
- Reuses Phase 4 UI patterns and the landing/preview path; adds a review-bundle assembler
  and CLI/web surface.
- Gives the human a single, evidence-backed feature-level review — the payoff of the
  envelope model.

## Open Questions

- Should the bundle be a generated sidecar export (durable, diffable) or assembled on
  demand from records? v1: assemble on demand; persist only if a durable feature-review
  record proves useful (aligns with ADR 0053 direction without overbuilding).
- How are very large aggregate diffs presented without overwhelming review? v1: per-child
  expandable diffs + an aggregate stat; full unified diff on request.
