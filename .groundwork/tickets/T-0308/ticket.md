---
id: T-0308
kind: ticket
node_type: leaf
work_type: technical_implementation
title: Model work types and actor routing metadata
status: done
assignee: null
requested_actor: null
priority: null
labels: []
parent: E-0004
depends_on: []
created_at: ""
updated_at: ""
---

## Problem

_No description recorded._

## Acceptance Criteria

- Tickets support a work_type field used by planner SOPs, policy, scheduler, validation, and context assembly.
- Tickets may optionally request an actor or declare required actor capabilities.
- Ticket Markdown export includes work_type and requested_actor.
- Work type remains organization-defined operational metadata, not a hardcoded SDLC status list.
