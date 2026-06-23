---
id: T-1034
kind: ticket
node_type: leaf
work_type: documentation
title: Update workflow & CLI docs for the human-CLI surface
status: done
assignee: null
requested_actor: null
priority: 0.6
labels: []
parent: G-0001
depends_on: []
created_at: "2026-06-23T20:46:20Z"
updated_at: "2026-06-23T21:48:57Z"
---

## Problem

Distill the T-1022 epic into canon: the ideal loop now uses gw next / gw ticket list --ready (not 'list --status todo', which ignores deps) and guided gw ticket claim, plus land --preview and edit --parent. Update WORKFLOW.md (loop steps), docs/reference/self-hosting.md (runbook), docs/contracts/cli.md (command surface), and .claude/skills/gw/SKILL.md (the 'no single claim command' note is now false). Leave ADRs as immutable records.

## Acceptance Criteria

- WORKFLOW.md loop references gw next / list --ready / claim instead of 'list --status todo' + bare manual transitions
- cli.md contract lists gw next, gw ticket claim, and land --preview; self-hosting runbook and gw SKILL.md reflect claim/next
