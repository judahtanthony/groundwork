---
id: T-1073
kind: ticket
node_type: leaf
work_type: technical_implementation
title: Land-preview API endpoint
status: done
assignee: human.owner
requested_actor: null
priority: 0.9
labels:
    - web-ui
    - operator-unblock
parent: T-1061
depends_on:
    - T-1064
created_at: "2026-06-24T23:06:25Z"
updated_at: "2026-06-24T23:29:44Z"
---

## Problem

Expose the staged landing diff over the coordinator API so the operator UI can preview a landing before approving (the server is the source of truth; the CLI's previewLanding is process-local). Add GET /api/v1/tickets/{id}/land/preview returning the staged change set via the server's git repo (s.repo.StagedDiff / HasStagedChanges), mirroring 'gw ticket land --preview' semantics, with the standard JSON error envelope. Backend prerequisite for the Land preview UI (T-1065).

## Acceptance Criteria

- GET /api/v1/tickets/{id}/land/preview returns the staged diff (and a staged=false signal when nothing is staged), matching gw ticket land --preview.
- Returns a not_a_repo / git_error JSON envelope when the project root is not a git work tree or git fails.
- Endpoint is read-only: it does not mutate the index or open an approval.
