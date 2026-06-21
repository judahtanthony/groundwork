# ADR 0035: Context-Miss Capture for the Dogfooding Feedback Loop

Status: Accepted

## Context

[ADR 0013](0013-canon-as-memory.md) frames `gw ticket context` as the read side of a
loop whose write side is distillation into canon, and names the dogfooding validator:
"whatever agents keep grepping for despite the brief is what the brief — and thus
canon — is missing." M3 is the first time real work runs through Groundwork, so it is
the first opportunity to observe those misses — but there is no mechanism to capture
them, and without capture the feedback loop is asserted, not operating.

## Decision

**A minimal, records-only context-miss capture.** While working a node, the worker
(a human in M3) records a *miss* against that node — what they needed that the brief
lacked — appended to the node's ephemeral journal / an ignored log (for example
`gw ticket context <id> --miss "<note>"`). Misses are ignored runtime state, not
canon; they carry no schema beyond a note, node id, and timestamp.

The value is in the **review/promotion step**: recurring misses are periodically
reviewed and promoted into the appropriate canonical home — an SOP under
`.groundwork/sops/`, a doc or ADR, or the context-brief assembly itself — closing the
loop. Promotion is a human act in v1, consistent with [ADR 0013](0013-canon-as-memory.md)'s
typed-promotion bloat filter ("does this change a canonical document?").

## Consequences

The canon-as-memory loop gains an observable signal in M3 without adding durable
bloat: misses stay ignored, and only the resulting canon edits are committed. It
validates [ADR 0013](0013-canon-as-memory.md) empirically — the docs ticket (T-1002)
is expected to surface at least one real miss that becomes a brief/SOP improvement
(T-1008). As later phases add AI executors, the same capture lets agents flag misses
into the same review/promotion step; the mechanism does not change.
