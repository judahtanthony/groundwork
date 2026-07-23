---
id: T-1105
kind: ticket
node_type: null
work_type: test_implementation
title: Add unit tests for internal/doctor checks; make doctor.Run read-only (no migrate)
status: backlog
assignee: null
requested_actor: null
priority: null
labels:
    - workflow
parent: T-1074
depends_on: []
created_at: "2026-07-23T12:22:43Z"
updated_at: "2026-07-23T12:22:43Z"
---

## Problem

T-1046 extracted gw doctor checks into a shared internal/doctor package (now used by both the CLI and the server settings endpoint), but it has no direct unit tests — the warn/error branches (actor load failure, DB migrate failure, config warnings) are unexercised. Add a table test covering ok/warn/error outcomes. Also: doctor.Run opens a second SQLite connection and calls db.Migrate() (pre-existing gw doctor behavior, faithfully preserved), so POST /api/v1/doctor is not strictly read-only; consider a read-only health check that does not apply pending migrations. Ref internal/doctor/doctor.go:35-84.

## Acceptance Criteria

- internal/doctor has a table test covering ok/warn/error outcomes for its checks
- The doctor health endpoint does not apply DB migrations as a side effect (or this is explicitly documented as intended)
