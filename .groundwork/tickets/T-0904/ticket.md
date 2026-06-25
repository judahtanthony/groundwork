---
id: T-0904
kind: ticket
node_type: leaf
work_type: technical_implementation
title: Resume interrupted runs from last checkpoint
status: backlog
assignee: null
requested_actor: null
priority: null
labels: []
parent: T-1071
depends_on:
    - T-0504
created_at: ""
updated_at: "2026-06-25T01:24:34Z"
---

## Problem

_No description recorded._

## Acceptance Criteria

- An interrupted run resumes from its last worktree checkpoint commit.
- In-flight work is not lost when a run is interrupted.
