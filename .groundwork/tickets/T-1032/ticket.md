---
id: T-1032
kind: ticket
node_type: leaf
work_type: technical_implementation
title: Make import->export byte-stable for the empty-timestamp convention
status: backlog
assignee: null
requested_actor: null
priority: 0.65
labels:
    - cli-ux
    - determinism
parent: T-1022
depends_on: []
created_at: "2026-06-22T19:28:34Z"
updated_at: "2026-06-22T19:28:34Z"
---

## Problem

Cold-start 'gw ticket import' populates concrete created_at/updated_at on bootstrap tickets that are committed with the empty-timestamp convention (created_at: ""), so a subsequent full 'gw ticket export' rewrites all such tickets' timestamps and breaks byte-stable round-trip (ADR 0020/0021). Repro: fresh checkout -> gw ticket import -> gw ticket export -> 50+ tickets show spurious created_at/updated_at diffs. Export should preserve the empty-timestamp convention (or import should not synthesize timestamps for empties), so import->export is a no-op on unchanged tickets. Surfaced while creating the T-1022 epic.

## Acceptance Criteria

_None recorded._
