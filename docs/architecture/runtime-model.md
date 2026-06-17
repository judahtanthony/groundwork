# Runtime Model

Groundwork runtime is centered on the coordinator.

## Process Roles

- `gw` CLI: primary human and script interface.
- `gw server`: localhost-only coordinator and dashboard.
- Agent runner: supervised process that runs one node attempt.
- Codex app-server: first supported coding-agent runtime.

## Run Modes

A run is one of two modes:

- **Planning run** (decomposition): triages a composite node, records research/design/requirements, and proposes child nodes and dependency edges. It produces a proposal, needs minimal or no worktree, and lands the node in `review`.
- **Implementation run**: executes a leaf in an isolated worktree and produces code.

## Runtime Flow

1. A node becomes eligible (`todo` and all dependencies satisfied).
2. Coordinator claims it transactionally in SQLite.
3. Coordinator creates a run and lease.
4. The agent receives a bounded context brief (`gw context`): ancestor spine, parent contract, direct dependencies, relevant SOPs, open escalations.
5. The claiming agent triages the node as leaf or composite.
6. Composite -> planning run produces a decomposition proposal -> `review`.
7. Leaf -> implementation run in an isolated worktree, checkpointing WIP as it goes.
8. Events stream to SQLite and local JSONL logs.
9. Approval requests pause gated actions (including `decompose`).
10. Validation results are recorded.
11. Landing is gated by validation and policy; WIP checkpoints are squashed into the landing commit, and ratified design is distilled into canon.

## Checkpoints And Distillation

An implementation run periodically commits work-in-progress as a **checkpoint** on its worktree branch so recovery can resume from the last checkpoint; checkpoints are squashed at landing and never reach `main` (see [ADR 0015](../adr/0015-run-checkpoints-squashed-at-landing.md)). At the ratification gate, durable design discovered during the run is **distilled into canon** rather than left in the ignored journal (see [ADR 0013](../adr/0013-canon-as-memory.md)).

## Defaults

- Sandbox posture: `workspace-write` by default; `read-only` and `danger-full` are selectable per run.
- Max concurrent agents: 4.
- Lease TTL: 90s; renewal (heartbeat) interval: 30s.

## Runtime State

Runtime state is local and ignored by default. It can be reconstructed enough for stable continuation, but exact model-internal state is not guaranteed after crashes.

