---
id: T-1020
kind: ticket
node_type: null
work_type: technical_implementation
title: Auto-rebuild (or hint) the store on cold-start CLI reads, not just gw server
status: backlog
assignee: null
requested_actor: null
priority: 0.1
labels: []
parent: G-0001
depends_on: []
created_at: "2026-06-22T12:38:07Z"
updated_at: "2026-06-23T20:44:08Z"
---

## Problem

Cold-start auto-import is wired only into 'gw server' startup (internal/cli/server.go:50: HasTickets() -> importExports). Plain CLI read commands (gw ticket tree, gw status, gw board) open the freshly-migrated empty store and report 'No tickets' instead of rebuilding from the committed exports under .groundwork/tickets/ or hinting to run 'gw ticket import'. On a fresh checkout (state.sqlite is git-ignored, so it never travels) a human running a read first hits a confusing empty tree. Extend the T-0902 cold-start rebuild to cover the CLI read path, or at minimum print a clear 'store empty; run gw ticket import' hint. Surfaced while dogfooding T-1005.

## Acceptance Criteria

- A read command (e.g. gw ticket tree) on a cold start with committed exports present shows the tree, or prints a clear hint to run 'gw ticket import'.
- The empty 'No tickets' output never appears when .groundwork/tickets/ contains exports.
