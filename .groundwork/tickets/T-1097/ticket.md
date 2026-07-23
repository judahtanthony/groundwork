---
id: T-1097
kind: ticket
node_type: null
work_type: technical_implementation
title: Import loses multi-line acceptance criteria (export->import not lossless)
status: backlog
assignee: null
requested_actor: null
priority: null
labels: []
parent: T-1074
depends_on: []
created_at: "2026-07-16T00:06:38Z"
updated_at: "2026-07-16T00:06:38Z"
---

## Problem

gw ticket import parses wrapped multi-line acceptance bullets lossily — it drops the indented continuation lines — so a store rebuilt from files (rm state.sqlite && gw ticket import) diverges from the committed sidecars for any ticket with multi-line acceptance (observed on T-0506, T-0507, E-0012 + 2). The committed ticket.md files are authoritative and correct; the store is behind. Violates the ADR 0053 file<->store round-trip invariant. Repro: rebuild the store, boot gw server -> 'N ticket(s) diverge'. The exporter wraps long acceptance bullets; the importer must reassemble them. Related to T-1087 (startup divergence hardening) but a distinct import-parser defect.

## Acceptance Criteria

- gw ticket import round-trips wrapped multi-line acceptance bullets losslessly
- A rebuilt store matches its committed sidecars byte-for-byte (no boot divergence)
