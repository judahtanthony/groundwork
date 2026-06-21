---
id: T-0606
kind: ticket
node_type: leaf
work_type: technical_implementation
title: Add actor registry and actor-aware policy matching
status: done
assignee: null
requested_actor: null
priority: null
labels: []
parent: E-0007
depends_on: []
created_at: ""
updated_at: ""
---

## Problem

_No description recorded._

## Acceptance Criteria

- .groundwork/actors.yaml defines local human and AI actors.
- The default scaffold creates human.owner and ai.codex.default.
- Actors have type, roles, runtime/model config, and coarse capabilities.
- Policy rules can match actor id, actor type, roles, work_type, action, files, and risk class.
- Requested actors are treated as routing hints and still require policy authorization.
