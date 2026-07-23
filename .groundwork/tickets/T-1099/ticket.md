---
id: T-1099
kind: ticket
node_type: null
work_type: technical_implementation
title: Add a run progress watchdog (fail a stalled run with no output)
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

Observed in the dogfood: a codex run hung on a bad model for 11+ minutes (0 CPU, no output past the startup banner) with no detection. The coordinator should watchdog runs: if a run emits no events / makes no progress for N minutes (config), mark it failed, kill the child, and surface it — instead of appearing 'running' indefinitely.

## Acceptance Criteria

- A run producing no events/progress for a configurable timeout is failed and its process killed
