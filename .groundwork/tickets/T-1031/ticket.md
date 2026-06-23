---
id: T-1031
kind: ticket
node_type: leaf
work_type: technical_implementation
title: Add gw ticket land --preview to show the diff to be landed
status: done
assignee: null
requested_actor: null
priority: 0.5
labels:
    - cli-ux
parent: T-1022
depends_on:
    - T-1023
created_at: "2026-06-22T15:39:30Z"
updated_at: "2026-06-23T19:43:19Z"
---

## Problem

gw ticket land <id> --preview (dry-run): show the staged change set / diff that the land_to_main gate would commit, without opening the approval. Helps the human review before landing.

## Acceptance Criteria

- gw ticket land --preview shows the staged diff the gate would commit, without opening an approval or contacting the coordinator
- Clear message when nothing is staged; --json parity
