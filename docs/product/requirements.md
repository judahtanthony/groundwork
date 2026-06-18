# Product Requirements

## V1 Requirements

- Provide a `gw` CLI as the primary interface.
- Store managed-project state under `.groundwork/`.
- Use SQLite as the local operational store during runtime.
- Ignore SQLite by default and export durable application state to readable files.
- Support a localhost-only single-user server and dashboard.
- Manage a uniform work-node tree, runs, approvals, validation, and recovery state.
- Support organization-defined `work_type` metadata on work nodes so planner SOPs can encode SDLC shape without expanding the status model.
- Triage each claimed node as leaf (execute) or composite (decompose) at claim time.
- Decompose composite nodes just-in-time into children plus a parent contract, gated as a reviewable proposal.
- Support dependency edges (a DAG overlay) so nodes dispatch only when prerequisites are satisfied.
- Propagate revisions upward via escalation; re-plan is human-gated in v1.
- Define local actors in `.groundwork/actors.yaml`, including a default human owner and default Codex AI actor, with actor type, roles, runtime/model settings, and coarse capabilities.
- Route claims, approvals, and gated actions through actor-aware policy matching while keeping v1 local and single-user.
- Use Codex as the first agent runtime while preserving an adapter boundary.
- Execute agent work in isolated worktrees.
- Require human approval before landing code to `main` and before accepting decomposition proposals in v1.
- Model landing and decomposition approval as policy gates so future autonomy can be enabled safely.
- Support validation templates by file type.
- Support work-type SOPs and per-action autonomy levels that loosen as SOPs, context, and validations mature.
- Represent risk as a 0–100 score mapped to named classes (`low`/`medium`/`high`/`critical`).
- Treat reversibility as a first-class gate input; force irreversible actions to `critical` regardless of score.
- Distill durable design into canon (committed files) at the ratification gate; have composite parents reconcile their children's promoted design so canon stays coherent and non-redundant.
- Assemble a bounded, node-specific context brief (`gw context`) at claim time from canon via the SQLite graph.
- Record actor identity and an actor configuration snapshot on runs so later audit remains stable after `.groundwork/actors.yaml` changes.
- Checkpoint run work-in-progress on the worktree branch for crash recovery, squashed into the landing commit so `main` stays clean.
- Support auto-approval for low-risk policy-defined actions such as internal documentation guidance updates.
- Provide startup and crash recovery from SQLite plus durable exports.

## Non-Goals For V1

- Multi-user LAN or hosted mode.
- Multi-repository orchestration.
- Mandatory external services such as Postgres, Redis, Temporal, Kubernetes, Docker, GitHub, Linear, Slack, or a hosted control plane.
- Complex SPA dashboard.
- Fully autonomous production deploys.
- Direct Markdown-to-SQLite live sync.
- Chat approvals.
- Reviewer-agent approvals.
- Multiple runtime adapters beyond Codex.

## Constraints

- Prefer open source libraries and open formats where possible.
- Use high-leverage local dependencies when they reduce implementation risk.
- Keep all required runtime services local to the process or filesystem.
- Keep docs and contracts current enough for future agents to implement without prior chat context.
