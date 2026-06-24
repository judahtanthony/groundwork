---
id: T-1038
kind: ticket
node_type: leaf
work_type: technical_implementation
title: Add gw binary-size guardrail (warn when the binary exceeds 100 MB)
status: backlog
assignee: null
requested_actor: null
priority: 0.85
labels:
    - web-ui
parent: T-1036
depends_on: []
created_at: "2026-06-24T15:37:14Z"
updated_at: "2026-06-24T15:37:14Z"
---

## Problem

The SPA will be embedded via go:embed, so the build must catch binary inflation: make build measures bin/gw and warns (does not fail) when it exceeds 100 MB. Baseline ~19 MB. Ships first as a standalone check.

## Acceptance Criteria

_None recorded._
