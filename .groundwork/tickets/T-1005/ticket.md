---
id: T-1005
kind: ticket
node_type: leaf
work_type: documentation
title: Update README with current runtime and development instructions
status: todo
assignee: null
requested_actor: null
priority: null
labels: []
parent: G-0001
depends_on: []
created_at: "2026-06-21T12:51:31Z"
updated_at: "2026-06-21T12:51:31Z"
---

## Problem

Refresh the root README for M1-M3: building gw (CGO_ENABLED=0 go build ./cmd/gw), gw init, importing the work tree, gw server, the self-hosting workflow (docs/reference/self-hosting.md), and the development gates (go vet, go test, -race, scripts/smoke.sh). Intended as low-risk documentation work to exercise the agent executor in Phase 4.

## Acceptance Criteria

- README explains how to build and run gw (server and CLI) using current commands.
- README documents the development workflow: build, test, -race, and smoke gates.
- README links to the self-hosting runbook and key architecture/contract docs.
