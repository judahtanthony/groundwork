---
id: T-0904
kind: ticket
node_type: leaf
work_type: technical_implementation
title: Resume interrupted runs from last checkpoint
status: done
assignee: null
requested_actor: null
priority: 0.4
labels: []
parent: T-1071
depends_on:
    - T-0504
    - T-1055
created_at: ""
updated_at: "2026-06-25T01:24:34Z"
---

## Problem

ADR 0015/0059: when a run is interrupted, a later run resumes from the run's last `gw/run/<run-id>` checkpoint commit (plus the resume packet from T-1055) rather than starting from node metadata alone, so in-flight work is not lost.

## Acceptance Criteria

- An interrupted run resumes from its last worktree checkpoint commit.
- In-flight work is not lost when a run is interrupted.
