# Architectural Proposals

This directory holds non-binding architectural proposals for future Groundwork work.
They are intentionally separate from accepted ADRs. Use them as source material for
future coding-agent sessions, then promote vetted decisions into ADRs, contracts,
architecture docs, policies, SOPs, and work-tree tickets.

## Status

All documents in this directory are **Draft** unless their header says otherwise. A
draft proposal records direction, goals, tradeoffs, and candidate low-level design. It
does not override existing ADRs, contracts, or project instructions.

## Proposal Set

Read in this order:

1. [Scalable Human-Agent Collaboration](0001-scalable-human-agent-collaboration.md)
2. [Hierarchical Planning And Approval Envelopes](0002-hierarchical-planning-and-approval-envelopes.md)
3. [Node Branching And Parent Integration](0003-node-branching-and-parent-integration.md)
4. [Resource Scope, Ownership, And Conflict Policy](0004-resource-scope-ownership-and-conflict-policy.md)
5. [Hierarchical Memory And Context Compression](0005-hierarchical-memory-and-context-compression.md)
6. [Role-Aware Actors And Local Identity](0006-role-aware-actors-and-local-identity.md)
7. [Full-SDLC Work Modeling](0007-full-sdlc-work-modeling.md)
8. [Agentic Software Factory Direction](0008-agentic-software-factory-direction.md)

## How To Use These Proposals

For each proposal:

1. Check whether it conflicts with accepted ADRs or contracts.
2. Split durable decisions into one or more ADRs.
3. Update the relevant architecture and contract docs.
4. Create Groundwork tickets small enough to validate independently.
5. Keep implementation behind policy gates until validation, risk, and review behavior
   are explicit.
