---
id: T-1100
kind: ticket
node_type: null
work_type: technical_implementation
title: buildPrompt must not surface system recovery_needed flags as the agent task
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

Observed in the dogfood: T-1039 got a recovery_needed decision (from the boot divergence check, root cause T-1097 import round-trip bug). resume.Assemble/buildPrompt then injected it into the agent prompt as 'Recommended next step: resolve blocker: recovery_needed: SQLite state diverged...', misdirecting the coding agent away from its actual task (scaffold the SPA). System/operator recovery flags must not be presented to the agent as work. Related: T-1097.

## Acceptance Criteria

- The agent prompt's open-questions/blockers exclude system-generated recovery_needed records
