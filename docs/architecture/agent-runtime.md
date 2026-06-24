# Agent Runtime

The first runtime target is Codex app-server. Groundwork should preserve an adapter interface for later runtimes.

## Codex Runtime Responsibilities

The Codex adapter should follow the responsibilities described conceptually by OpenAI Symphony:

- prepare a per-node workspace (planning runs may use a minimal or no worktree; implementation runs use an isolated worktree),
- construct a prompt from workflow, work-type SOPs, actor instructions/capabilities, the ancestor/contract context, and node context,
- launch the coding-agent app-server,
- stream updates,
- triage the node and, for composite nodes, produce a decomposition proposal (children, parent contract, dependency edges) rather than code,
- handle approvals and input-required events, including the `decompose` gate and escalation/upward-revision,
- write durable ticket-attached request/decision records before exiting on any blocker
  that must survive rebuild,
- record observability data,
- retry or pause according to coordinator policy.

The runtime receives an actor configuration chosen by the coordinator. For the default Codex actor this includes runtime `codex`, model selection, sandbox posture, and any configured instructions, tools, skills, or MCPs. Runs must record the actor id and a snapshot of that configuration so historical runs remain auditable after actor definitions change.

## Workspace Safety

Agents must run only inside the run worktree under `.groundwork/worktrees/`. The runtime must validate path containment before launch.

## Event Streaming

Runtime events should stream to SQLite and to `.groundwork/runs/<run-id>/events.ndjson`. A human-readable `transcript.md` should be generated locally but ignored by default.

## Pause, Block, And Resume

Pause should stop at a safe boundary where possible. A blocked autonomous run should not
remain alive waiting for human input. It checkpoints work when applicable, writes events
and runtime evidence, exports the durable blocker/request record, optionally creates a
dependent decision ticket, moves the original ticket to an explainable blocked state,
releases its lease, and exits.

Resume starts a new turn from a structured packet assembled from durable ticket/run/canon
state: current ticket context, ancestor contract, acceptance criteria, dependencies,
resolved decision/input records, rework notes, handoff summary, checkpoint/diff refs,
validation state, artifacts, and relevant transcript excerpts only.

## Approvals

When Codex requests a dangerous action, operator input, or a policy-gated capability,
Groundwork should create a durable ticket-attached request record when the wait must
survive rebuild, then project it into the live approval/input queue until a decision is
made.
