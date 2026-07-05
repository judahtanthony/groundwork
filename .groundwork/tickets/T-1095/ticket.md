---
id: T-1095
kind: ticket
node_type: null
work_type: technical_design
title: Multi-provider agent roles by SDLC stage (Codex/Claude Code/Gemini)
status: backlog
assignee: null
requested_actor: null
priority: null
labels: []
parent: T-1074
depends_on: []
created_at: "2026-07-05T17:56:08Z"
updated_at: "2026-07-05T17:56:08Z"
---

## Problem

Support multiple agent providers routed by role/SDLC stage (ADR 0048/0055 role-aware actors). Initial intent: Codex=triage/planning, Claude Code=code implementation, Gemini=review/judging; expand to PRDs, designs, etc. as needed. SOPs must stay provider-agnostic so any provider can execute a work type.

## Acceptance Criteria

- Actor registry + trust policy can route work types to different providers
- Codex handles triage/planning, Claude Code implementation, Gemini review/judging (initial mapping)
