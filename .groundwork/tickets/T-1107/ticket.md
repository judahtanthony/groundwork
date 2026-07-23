---
id: T-1107
kind: ticket
node_type: null
work_type: technical_implementation
title: 'land_to_main aborts: approval decision-record write dirties tree before git checkout main'
status: backlog
assignee: null
requested_actor: null
priority: null
labels:
    - workflow
parent: T-1074
depends_on: []
created_at: "2026-07-23T13:59:37Z"
updated_at: "2026-07-23T13:59:37Z"
---

## Problem

During gw ticket land <root> (land_to_main), approving the gate writes a rework/approval decision record to .groundwork/tickets/<id>/decisions.ndjson in the same working tree, AFTER the landing staged its files. mergeRootToMain then runs 'git checkout main' which git refuses: 'Your local changes to decisions.ndjson would be overwritten by checkout.' Result: the node is recorded landed (status=done, export committed on the integration branch) but the merge aborts, leaving main un-updated and requiring manual recovery. Repro: land any root to main. Fix options: commit/stage the decision record as part of the landing before checkout, write it after the merge, or checkout with the ticket sidecars stashed. Ref internal/server/landing_commit.go (finishLanding: ratify writes decision, then checkoutRootIntegrationBranch/mergeRootToMain).

## Acceptance Criteria

- land_to_main completes the git merge without a manual 'commit decisions.ndjson then re-run' step; the decision record does not block git checkout
