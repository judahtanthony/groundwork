---
id: T-1019
kind: ticket
node_type: null
work_type: technical_implementation
title: Add gw command to install the gw skill into a repo
status: backlog
assignee: null
requested_actor: null
priority: null
labels:
    - tooling
    - agent-integration
parent: T-1022
depends_on: []
created_at: ""
updated_at: "2026-06-23T20:02:03Z"
---

## Problem

Add a gw CLI command (e.g. `gw skill install`) that writes the Groundwork agent skill into the target repo's .claude/skills/gw/SKILL.md so Claude Code (and compatible agents) can discover and drive the gw CLI. The skill currently lives in this repo's .claude/skills/gw/; this command makes installing it reproducible in any managed project instead of copying by hand.

## Acceptance Criteria

- A gw subcommand installs the gw skill to .claude/skills/gw/SKILL.md in the current repo
- The installed skill content is sourced from a canonical embedded/templated copy, not hand-duplicated
- Command is idempotent and does not clobber local edits without an explicit overwrite flag
- Surfaced in 'gw help' / 'gw skill -h' with usage
