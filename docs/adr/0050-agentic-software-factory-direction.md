# ADR 0050: Agentic Software Factory Direction

Status: Draft
Implemented: Partial

## Implementation State

Groundwork already has the local-first control-plane foundation: work tree, coordinator, gates, validation records, actors, canon, and planned Codex runtime direction. The broader software-factory model remains draft future direction and does not change the active Phase 4 operator-UI boundary or the Phase 5 bounded-autonomy boundary before Phase 6 runtime work.

## Goal

Groundwork should evolve into a local-first agentic software-factory control plane:
a system where humans define goals, contracts, policies, and review thresholds while
agents perform structured background work in isolated environments with validation,
audit, memory, and progressively loosenable authority.

This draft ADR records industry direction and turns it into a product evolution path for
Groundwork. It is not an accepted ADR and does not override the current Phase 4
operator-UI boundary, Phase 5 bounded-autonomy work, or Phase 6 Codex runtime boundary.

## Industry Direction

Public examples point in the same direction: coding agents are moving from interactive
assistants to background workers governed by harnesses, sandboxes, deterministic checks,
and review systems.

Representative examples:

- Ramp Inspect is an internal background coding agent focused on closing the loop from
  code generation through verification. Ramp's public writeup emphasizes a background
  agent that writes code and verifies it rather than stopping at a patch suggestion:
  <https://builders.ramp.com/post/why-we-built-our-background-agent>.
- Modal describes Ramp Inspect as using fast full-stack sandbox environments and reports
  that the system gives builders access to high-parallelism engineering capacity grounded
  in internal code, context, and tooling:
  <https://modal.com/blog/how-ramp-built-a-full-context-background-coding-agent-on-modal>.
- OpenInspect / background-agents is an open-source system inspired by Ramp Inspect. Its
  architecture includes background sessions, isolated environments, GitHub/Linear/Slack
  style triggers, and a single-tenant security model:
  <https://github.com/ColeMurray/background-agents>.
- Stripe Minions are described by Stripe as one-shot, end-to-end coding agents, with
  follow-up material discussing implementation details:
  <https://stripe.dev/blog/minions-stripes-one-shot-end-to-end-coding-agents-part-2>.
  Public coverage highlights "blueprints" that mix deterministic routines with agent
  work so workflows stay efficient while retaining adaptability:
  <https://www.infoq.com/news/2026/03/stripe-autonomous-coding-agents/>.
- OpenAI's harness-engineering framing says the work is not only the model; it is the
  environment, tools, repository legibility, constraints, validation, and feedback loops
  around the agent:
  <https://openai.com/index/harness-engineering/>.
- Anthropic's agent-building guidance favors simple composable workflows where possible,
  with tool quality, context control, routing, parallelization, and evaluator-style loops
  used deliberately:
  <https://www.anthropic.com/engineering/building-effective-agents>.

The common pattern is:

```text
human request or system event
  -> governed work intake
  -> planner / router
  -> isolated agent run
  -> deterministic validation
  -> agent or human review
  -> branch / PR / commit
  -> audit, memory, and policy update
```

## Product Positioning

Groundwork should not compete first on cloud sandbox infrastructure. Ramp, Modal, Ona,
and similar systems show that scalable sandbox execution is important, but Groundwork's
current differentiated advantage is the local-first work control plane:

- uniform work tree,
- hierarchy and dependency graph,
- SOPs and work types,
- actor registry,
- approval gates,
- validation policy,
- canon/memory loop,
- deterministic ticket exports,
- local operational state.

The long-term promise is:

```text
Groundwork is a local-first control plane for agentic software work. It turns goals into
governed work, runs agents inside explicit contracts, validates their output, records
what happened, and progressively reduces human review load as trust is earned.
```

## Software Factory Model

Groundwork should model software work as a factory with visible control points:

1. **Intake**: human, CLI, web admin, issue, webhook, monitor, or scheduled trigger
   creates or updates a root node.
2. **Planning**: planner converts the goal into a contract, decomposition, dependencies,
   resource scopes, validation expectations, and actor candidates.
3. **Authorization**: policy decides what may run, which role must approve, and what
   escalation triggers apply.
4. **Execution**: background agents run in isolated worktrees or sandboxes.
5. **Observation**: events, transcripts, tool calls, costs, validation, and diffs stream
   into run records.
6. **Review**: reviewer agents and humans inspect summaries, exceptions, and final
   outcomes.
7. **Integration**: parent nodes integrate child outputs into parent/root branches.
8. **Landing**: root outcomes land to the project branch through policy gates.
9. **Learning**: summaries, context misses, SOP updates, policy suggestions, and canon
   changes feed the next runs.

## Harnesses And Blueprints

Groundwork SOPs should mature into executable harness blueprints. A blueprint is a
workflow made from deterministic steps plus agent steps.

Example blueprint shape:

```yaml
blueprint:
  work_type: dependency_upgrade
  steps:
    - id: prepare
      kind: deterministic
      command: gw context assemble
    - id: edit
      kind: agent
      runtime: codex
      prompt_from:
        - node_brief
        - parent_contract
        - sop
        - scope
    - id: fast_validation
      kind: deterministic
      commands:
        - go test ./...
    - id: review
      kind: agent_review
      required_when:
        - diff_lines_over: 200
        - touched_unexpected_files: true
    - id: land_to_parent
      kind: policy_gate
      action: land_to_parent
```

This gives Groundwork the same design shape as industry software factories:
deterministic routines do what code can do reliably, while agents handle judgment and
adaptation inside bounded steps.

## Background Agent Infrastructure

The local-first MVP can start with local worktrees. The architecture should keep room for
additional execution backends:

- local Codex worktree,
- local container or VM sandbox,
- remote sandbox provider,
- cloud agent runner,
- specialized hosted integration runner.

The runtime interface should preserve these responsibilities:

- prepare workspace,
- inject context,
- enforce path and scope containment,
- run agent,
- stream events,
- record transcript and tool calls,
- checkpoint work,
- produce result ref or patch,
- summarize,
- report validation and cost metadata.

Groundwork should own orchestration and governance even when execution moves out of the
local process.

## Governance Model

The system should reduce human review load without removing accountability. It should do
that by layering control:

- approved planning envelopes,
- resource scopes and ownership,
- deterministic validation,
- reviewer-agent checks,
- batch approvals,
- sampling of auto-approved work,
- policy suggestions instead of self-elevation,
- circuit breakers that demote work types or actors after bad outcomes.

Autonomy should be earned per work type and per actor, not granted globally.

## Recommended Evolution Path

### 1. Make Phase 4 A Real Harness

The Codex runtime should produce durable run records, not just launch an agent:

- isolated worktree,
- event stream,
- transcript,
- checkpoint refs,
- touched-file capture,
- validation records,
- completion summary,
- resume context,
- result ref or patch.

### 2. Convert SOPs Into Blueprint Seeds

Keep SOPs human-readable, but add structured sections that can drive workflows:

- required inputs,
- deterministic setup steps,
- agent prompt sources,
- validation commands,
- expected outputs,
- escalation triggers,
- review checklist.

### 3. Add Approval Envelopes And Parent Integration

Approve parent/root contracts. Let child work proceed inside approved limits and land to
parent integration targets when policy allows. Keep root landing human-gated until the
system has enough validation and trust data.

### 4. Add Resource Scope And Ownership

Use planned file/resource scope to prevent conflicting parallelism and to route approvals
like CODEOWNERS. Compare actual diffs against planned scope at review and landing.

### 5. Make Memory Hierarchical

Require completion summaries. Feed direct dependency summaries and parent memory into
child context. Keep raw transcripts available for audit, but do not make every agent read
every sibling transcript.

### 6. Build The Web Admin As A Governance Surface

The web admin should not be only a board. It should make non-CLI users productive by
showing:

- active background runs,
- root branches and integration status,
- approvals by required role,
- exceptions grouped by parent,
- planned vs actual scope,
- validation state,
- reviewer-agent findings,
- final root outcome review.

### 7. Add Reviewer Agents Before Auto-Landing

Reviewer agents can check contract compliance, scope drift, test quality, and summary
accuracy before the human reviews. Mature low-risk work can then move to batch approval,
sampling, and eventually policy-approved child landing.

### 8. Start With Boring High-ROI Work

The first agentic automation domains should be repetitive and verifiable:

- documentation updates,
- test generation,
- dependency upgrades,
- lint and formatting cleanup,
- mechanical refactors,
- flaky-test investigation,
- config cleanup,
- narrow migration preparation.

Feature development should use approval envelopes and parent integration until the system
has enough track record.

## Low-Level Design Implications

Future ADRs and tickets should consider these schema and API additions:

- `run_events` with richer event types for tool calls, validation, review, scope drift,
  cost, and summaries.
- `node_results` or equivalent records for result refs, patches, touched files, and
  integration status.
- `approval_envelopes` associated with root/composite nodes.
- `resource_scopes` in ticket exports or parent contracts.
- `actor_roles` and local identity config.
- `blueprint` metadata under SOPs or work-type config.
- `parent_memory` or synthesized summaries for context assembly.
- scheduler checks for active resource-scope conflicts.
- policy matcher inputs for planned scope, actual scope, role requirements, validation,
  branch target, and envelope compliance.
- web/admin APIs for role-aware queues, exceptions, integration status, and run traces.

## Open Questions

- Should blueprints live under `.groundwork/sops/`, `.groundwork/blueprints/`, or be
  embedded in work-type policy?
- What is the smallest Phase 4 harness that proves the model without building a full
  remote sandbox platform?
- Should background triggers be delayed until after local Codex runs can safely resume
  and summarize?
- How should Groundwork interoperate with GitHub PRs once root branches exist?
- Which metrics should govern earned autonomy: validation pass rate, revert rate,
  review findings, scope drift, or sampled review quality?

## Candidate Follow-Up ADRs

- Groundwork as a local-first agentic software-factory control plane.
- SOP blueprints as executable harness definitions.
- Runtime backends as interchangeable execution environments.
- Reviewer agents and sampling before autonomous landing.
- Background triggers as governed work intake.
