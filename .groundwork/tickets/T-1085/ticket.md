---
id: T-1085
kind: ticket
node_type: leaf
work_type: technical_implementation
title: Review-bundle assembler and gw review bundle CLI
status: done
assignee: human.owner
requested_actor: null
priority: 0.8
labels:
    - phase-5
parent: T-1070
depends_on:
    - T-1084
created_at: "2026-06-25T20:23:30Z"
updated_at: "2026-06-26T01:33:04Z"
---

## Problem

Deterministically assemble a parent/root review bundle (summaries, diffs, validation, findings, exceptions, recommendation); expose gw review bundle (ADR 0057).

## Acceptance Criteria

- gw review bundle <node> --json returns the bundle assembled from subtree records.
