---
id: T-1098
kind: ticket
node_type: null
work_type: technical_implementation
title: gw run cancel must terminate the child agent process
status: backlog
assignee: null
requested_actor: null
priority: null
labels: []
parent: T-1074
depends_on: []
created_at: "2026-07-16T00:38:33Z"
updated_at: "2026-07-16T00:38:33Z"
---

## Problem

Observed in the T-1036 dogfood: 'gw run cancel R-0001' set the run to cancelled in the store but left the codex OS processes (node wrapper + binary) alive and hung; they had to be pkill'd by hand. Run cancel/interrupt must propagate a kill to the launched child process (and its process group), so a cancelled/stuck run actually stops.

## Acceptance Criteria

- Cancelling a run kills its codex/agent OS process, not just the store record
