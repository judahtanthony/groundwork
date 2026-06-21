# ADR 0039: Value/Prioritization v1 — Hierarchical Sibling-Scoped Priority

Status: Accepted

First concrete slice of the value/prioritization layer ([ADR 0036](0036-work-as-universal-substrate.md) Layer 1).

## Context

[ADR 0036](0036-work-as-universal-substrate.md) names the value/prioritization layer
(Layer 1) as having **no substrate today**: the scheduler picks among eligible nodes
FIFO-by-id, and `priority`/`risk_score` are inert display fields. We need a minimal
value signal that orders the eligible set, is human-settable now, and is a clean seam
for the eventual multi-signal model (value, effort, complexity, risk, confidence,
tree depth, …) — without violating the architectural invariants
([ADR 0037](0037-transitional-defaults-vs-invariants.md)): default-deny, deterministic
export, and the single serialized coordinator.

## Decision

**A node carries an optional `priority` float in `[0,1]`, default `0`, used only to
order the eligible set — never to authorize.** Scoring (which eligible node runs first)
is separated from gating (whether a node may be claimed at all); priority is purely the
former, so default-deny ([ADR 0028](0028-gate-evaluation-engine.md)) is untouched.

**Priority is sibling-scoped by construction.** The scheduler orders eligible nodes by
the **lexicographic priority path** root→node — each level's own `priority`, unset
defaulting to `0`, compared **descending** — with the root→node **id-path** as the
DFS/FIFO tiebreak (ascending). Because siblings share their ancestor prefix, a node's
own `priority` only ever discriminates it from its siblings; cross-subtree order is
resolved at the ancestor divergence point, not by a global scalar.

Properties that follow:

- **FIFO by default, explicit to jump.** Unprioritized work (all `0`) runs in DFS /
  creation order; setting any node `> 0` floats its **whole subtree** ahead of
  lower-priority siblings. No inheritance machinery — the path comparison supplies the
  hierarchy.
- **No write cascade.** Effective priority is the path computed at query time, so
  re-prioritizing an initiative reorders its subtree with **zero descendant writes**.
  Reordering one node is a single write.
- **Insertion-between** (re-plan, drag-n-drop): assign a midpoint between neighbors'
  values (`0.15` between `0.1` and `0.2`; a re-planned `0.15` node's children take
  `0.151…0.159`). Within sibling scope this rarely stresses `float64` precision; the
  escape hatch is a **local renormalization** of one sibling set, never a tree-wide
  event. Determinism holds ([ADR 0020](0020-canonical-encoding-deterministic-export.md)):
  shortest-round-trip float formatting is stable and the midpoint is deterministic.
- **Setting priority is the creator's / decomposer's act** — a human today, a *gated*
  agent action later. That is the Layer-1 judgment migrating along the trust gradient
  ([ADR 0038](0038-authority-as-loosenable-gate.md)).

**The ordering is a single `score(node)` seam.** v1 implements `score` as the priority
path above; the future multi-signal value model replaces the function without touching
the scheduler loop.

**Classification ([ADR 0037](0037-transitional-defaults-vs-invariants.md)):**
human-set priority and human-as-prioritizer are *transitional defaults* (loosenable);
deterministic ordering, default-deny, and single-coordinator arbitration are
*invariants* preserved.

## Consequences

The scheduler's eligible-set ordering changes from id to `score`; `priority` becomes a
live `[0,1]` float input (was inert) settable at create/decompose and carried through
export. Deferred to later decisions: the multi-signal value model, automated value
assessment, and the drag-n-drop prioritization UI (a dashboard surface) — the data
model here already supports the last (midpoint = one write). This ADR builds the seam,
not the full model.
