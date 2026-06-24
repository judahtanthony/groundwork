# State Model

Groundwork uses committed/exported files for durable application state and SQLite as a
transactional projection plus runtime coordination store.

## Which Store: The Master Test

For any datum, ask: **could Groundwork rebuild this from the files plus git, or is its loss irreversible?** SQLite may hold only what is recomputable or safely lost; anything whose loss is irreversible must be a file. A datum's home is set by the role it plays once it exists, not by where it was produced. State falls into three tiers (see [ADR 0012](../adr/0012-three-tier-state-and-ratification-timing.md)):

- **Ephemeral runtime** — SQLite plus ignored run logs. Recomputable or safely lost.
- **Durable operational records** — file-authoritative, projected into SQLite. Mutated
  through `gw`/the coordinator for validation and transactions, but durable success
  requires the filesystem source of truth or a durable replay record (ADR 0053).
- **Canonical knowledge** — file-authoritative, committed. The file is the source of truth; SQLite at most indexes it.

SQLite is the live index/graph; files hold durable content. This is what lets context be queried per-node (`gw context`) while design and operational memory remain rebuildable.

Model sessions are disposable. Run transcripts and events are evidence for audit and
debugging, not the authoritative memory for resuming work. If losing a pending request,
proposal, or blocker would strand a ticket or change the meaning of its state, that
record is durable operational state and must be exported with the ticket (ADR 0051).

## Canonical Knowledge

File-authoritative and committed:

- code,
- product, visual (`docs/design/`), and technical design (ADRs, architecture docs),
- trust, risk, validation, and autonomy policies,
- local actor registry (`.groundwork/actors.yaml`),
- work-type SOPs,
- distilled design decisions promoted to canon.

Distilled design is written to a file at the **ratification gate** — when a decision becomes binding on other work (a decomposition proposal accepted, a node lands, a policy/SOP change approved) — not continuously and not only at root completion. Before ratification it is mutable ephemeral state; a decision never ratified never touches the repo. See [ADR 0013](../adr/0013-canon-as-memory.md).

## Durable Application State

Commit or preserve (file-authoritative, projected into SQLite):

- node identity, title, description, acceptance criteria, labels, priority, status, assignee, advisory kind, work type, requested actor, and parent,
- parent design/contract for composite nodes,
- dependency edges between nodes,
- escalation / upward-revision events and re-plan decisions,
- ticket-attached decision, input, approval, rework, and recovery records that explain
  blocked/review/rework states,
- work-type SOPs and context,
- meaningful node timeline entries,
- workflow prompt and operating policy,
- trust, risk, and validation policies,
- decomposition, landing, and other gate decisions and human overrides, with live
  approval rows treated as rebuildable projections when a durable record is required,
- run actor ids and actor configuration snapshots,
- code commits.

## Runtime State

Ignore by default:

- SQLite database and WAL/SHM,
- active leases,
- process IDs,
- live approval queue handles whose durable semantic request is exported with the ticket,
- run transcripts and model events,
- raw command output,
- per-node journals of in-progress decision notes (until distilled to canon),
- worktree contents before landing (preserved in-flight by run checkpoints; see [ADR 0015](../adr/0015-run-checkpoints-squashed-at-landing.md)),
- generated dashboard views.

## Ticket Statuses

Groundwork adapts Symphony's states. Symphony uses `Backlog`, `Todo`, `In Progress`, `Human Review`, `Merging`, `Rework`, and `Done`. Groundwork uses lower-case machine states:

- `backlog`: captured but not eligible for agent work.
- `todo`: ready for dispatch; eligible only when all dependencies are satisfied.
- `in_progress`: claimed and being worked.
- `blocked`: waiting on approval, an unsatisfied dependency, an escalation/re-plan, conflict, or external input.
- `review`: a prepared implementation **or a decomposition proposal** awaiting human, policy, or future reviewer-agent review.
- `rework`: review failed and the agent should revise.
- `approved`: review passed and landing may proceed if validation and trust gates pass.
- `landing`: actively landing validated changes.
- `done`: completed.
- `cancelled`: terminal, intentionally stopped.

Differences from Symphony:

- `Human Review` becomes `review` because later review may be policy or reviewer-agent driven.
- `Merging` becomes `landing` because Groundwork may land through trunk commits, local branch merges, fast-forward, or future PR integration.
- `blocked` is first-class because tactical approvals are central.
- `backlog` remains non-dispatchable for ideas and future work.

A blocked ticket must have an explainable blocker: an unsatisfied dependency or a
durable ticket-attached request/decision/recovery record. Startup reconciliation should
surface `recovery_needed` when a status implies a blocker but the durable context needed
to recover it was never exported.
