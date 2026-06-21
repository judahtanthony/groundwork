# Product Vision

Groundwork is a local-first work coordination system for engineers who want to manage coding agents without adopting a hosted issue tracker, queue service, database service, or cloud control plane.

The target user for v1 is a solo developer or solopreneur running the whole system on one machine. The system should later grow toward small-team and integration use cases without compromising the local-first core.

## Long-Term Goal

Groundwork aims for a Symphony-style operating model as the near-term shape of a broader
destination (see [ADR 0036](../adr/0036-work-as-universal-substrate.md)):

- Humans manage work, constraints, trust, validation, and visibility.
- Agents increasingly complete tasks end-to-end.
- Human code contribution trends toward zero for well-understood, well-constrained work.
- The system records enough context for later agents to understand what happened and why.

Beyond that, Groundwork treats **work** as a scale-invariant substrate (requirements,
procedures, definition-of-done, and value), with software as the seed domain. Over time it
migrates not just execution but **judgment and authority** — prioritization (value),
direction (solutions and execution), and improvement (the system enhancing itself) — from
humans into AI agents, balancing velocity against quality. Human-in-the-loop is a
transitional state governed by complexity and risk, not a permanent design invariant: every
human-authority point is a configurable, retractable position on a trust gradient
([ADR 0037](../adr/0037-transitional-defaults-vs-invariants.md),
[ADR 0038](../adr/0038-authority-as-loosenable-gate.md)). The aim is to elevate human
attention to ever-higher altitude through reliable, sustainable delegation.

## Local-First Motivation

Groundwork should lower barriers by avoiding mandatory SaaS dependencies. A user should be able to clone a repo, run `gw`, and manage local agent work with state colocated beside the code.

The system should be transparent to humans and agents. Tickets, policies, workflow guidance, and durable decisions should be readable as committed files. Runtime state may use SQLite and local logs because active coordination needs reliable transactions.

## Product Identity

- Name: Groundwork.
- CLI: `gw`.
- Managed-project directory: `.groundwork/`.
- Implementation language: Go.
- First runtime: Codex.

## Success Definition

Groundwork succeeds when a fresh agent can enter a repo, read the durable docs and `.groundwork/` state, understand the current work, pick a safe node, triage it — decomposing it into children when needed or executing it in an isolated worktree when it is a leaf — request approvals when needed, escalate revisions upward, run validation, and present or land changes according to policy.

