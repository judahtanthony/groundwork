# Agent Runtime

The first runtime target is Codex app-server. Groundwork should preserve an adapter interface for later runtimes.

## Codex Runtime Responsibilities

The Codex adapter should follow the responsibilities described conceptually by OpenAI Symphony:

- prepare a per-node workspace (planning runs may use a minimal or no worktree; implementation runs use an isolated worktree),
- construct a prompt from workflow, task-type SOPs, the ancestor/contract context, and node context,
- launch the coding-agent app-server,
- stream updates,
- triage the node and, for composite nodes, produce a decomposition proposal (children, parent contract, dependency edges) rather than code,
- handle approvals and input-required events, including the `decompose` gate and escalation/upward-revision,
- record observability data,
- retry or pause according to coordinator policy.

## Workspace Safety

Agents must run only inside the run worktree under `.groundwork/worktrees/`. The runtime must validate path containment before launch.

## Event Streaming

Runtime events should stream to SQLite and to `.groundwork/runs/<run-id>/events.ndjson`. A human-readable `transcript.md` should be generated locally but ignored by default.

## Pause And Resume

Pause should stop at a safe boundary where possible. Resume starts a new turn with summarized prior context, current ticket state, current diff, validation state, unresolved tasks, and relevant transcript excerpts.

## Approvals

When Codex requests a dangerous action, operator input, or a policy-gated capability, Groundwork should create an approval record and pause the gated action until a decision is made.

