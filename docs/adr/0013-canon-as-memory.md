# ADR 0013: Canon As Memory

Status: Accepted
Implemented: Partial

## Context

Groundwork must "record enough context for later agents to understand what happened and why," but the rich causal record (reasoning, rejected alternatives, escalation causes) is produced inside runs whose transcripts are ephemeral and ignored (ADR 0007, ADR 0012). Committing that record verbatim would bloat the repo with low-value history and create merge conflicts across parallel worktrees. We need durable "why" without an exhaustive archive.

## Decision

**The canonical documents are the memory; we do not archive history.** A decision record's job is to justify a change to a canonical document; once that change is committed, the record has done its work and can be discarded. The durable footprint of a completed subtree is therefore a diff to canon, not a pile of transcripts.

- **Journal vs canon.** Each node keeps an append-only **journal** of decision notes — tier-1 ephemeral state (ADR 0012), one file per node, ignored by default, so parallel runs never conflict on it. **Canon** is the typed durable set: product design, visual design (`docs/design/`), technical design (ADRs/architecture), policy (`.groundwork/policies/`), and SOPs (`.groundwork/sops/`).
- **Typed promotion is the bloat filter.** The test for durability is "does this change a canonical document?" If yes, edit that document in place (replace, do not append, so repo size tracks the current design). If it names no canonical home, it is not worth keeping and stays in the journal.
- **Two cadences, split by direction.** *Promote-on-dependency* (forward): a decision other nodes depend on goes into canon immediately, via the parent **contract**, so siblings see it. *Distill-on-completion* (retrospective, lossy): outcomes and lessons roll up when a node closes and compact as they ascend — a leaf emits a few lines to its parent, a composite consolidates its children's lines, the root emits one small evolution note linking to canon diffs.
- **Parent reconciliation.** A composite node owns reconciling the design its children promote: it reviews and refines the promoted contributions so canon stays coherent and non-redundant — no conflicting or repetitive design description. Reconciliation happens at the parent, at ratification, serialized through the coordinator (not concurrently in worktrees), which also keeps canon writes conflict-free.
- **`gw context` is the read side.** The same canon that distillation writes is read back as a bounded, node-specific brief (ancestor spine + parent contract + direct dependencies + relevant SOPs + open escalations). Distillation writes canon at ratification; `gw context` reads canon at claim time. They are two halves of one loop.

## Consequences

`state-model.md`, `work-tree.md` (parent reconciliation), `run-logs.md` (journal), and `cli.md`/`http-api.md` (`gw context`) are updated. Repo growth is bounded to surviving, ratified, canonical changes. Dogfooding validates the loop: whatever agents keep grepping for despite the brief is what the brief — and thus canon — is missing.
