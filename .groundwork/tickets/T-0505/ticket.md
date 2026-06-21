---
id: T-0505
kind: ticket
node_type: leaf
work_type: technical_implementation
title: Record actor snapshots on runs
status: done
assignee: null
requested_actor: null
priority: null
labels: []
parent: E-0006
depends_on: []
created_at: ""
updated_at: ""
---

## Problem

_No description recorded._

## Acceptance Criteria

- Runs persist actor_id, runtime, model, and actor_snapshot_json.
- Actor snapshots preserve the selected actor configuration for audit after .groundwork/actors.yaml changes.
- Resume and run detail output include the actor identity and relevant snapshot summary.
