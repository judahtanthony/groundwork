---
id: T-1104
kind: ticket
node_type: null
work_type: test_implementation
title: Add a concurrent -race test for policy.Set lock (ReplaceTrust vs Evaluate/AuthorizeClaim)
status: backlog
assignee: null
requested_actor: null
priority: null
labels:
    - workflow
parent: T-1074
depends_on: []
created_at: "2026-07-23T01:05:33Z"
updated_at: "2026-07-23T01:05:33Z"
---

## Problem

T-1045 added a sync.RWMutex to policy.Set with ReplaceTrust() (write-lock swap) and RLock in gate Evaluate/AuthorizeClaim to support live trust-rule editing. The lock is correct by inspection and go test -race passes, but the existing tests are sequential so -race never actually stresses the lock. Add a goroutine test that loops ReplaceTrust concurrently with AuthorizeClaim/Evaluate under -race to lock in the guarantee. Also consider a lost-update guard on concurrent amend_policy approvals (second approval overwrites with possibly-stale content; acceptable for single-operator v1). Ref internal/policy/load.go:26, internal/server/policies.go:170.

## Acceptance Criteria

- A -race test runs ReplaceTrust concurrently with AuthorizeClaim/Evaluate and passes, exercising the policy.Set lock
