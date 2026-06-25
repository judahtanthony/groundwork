---
id: T-1072
kind: ticket
node_type: leaf
work_type: technical_implementation
title: 'Operator UI: shell and navigation foundation'
status: done
assignee: human.owner
requested_actor: null
priority: 0.94
labels:
    - web-ui
    - operator-unblock
parent: T-1061
depends_on: []
created_at: "2026-06-24T23:06:25Z"
updated_at: "2026-06-24T23:18:05Z"
---

## Problem

Extract the shared server-rendered layout (sidebar nav, topbar/breadcrumb, SSE live-refresh script) from the existing dashboard into a reusable base template, and establish the multi-page route pattern that the operator screens build on. Activate the Tickets and Approvals navigation entries (remove the 'soon' placeholders). Server-rendered Go html/template + SSE only; no SPA/Vite (ADR 0042 progressive complexity; design README). Dependency root for all operator-unblock screens.

## Acceptance Criteria

- A shared base layout (sidebar, topbar, SSE refresh) is reused by the dashboard and new operator pages instead of being duplicated.
- Tickets and Approvals nav entries are active links (no 'soon' pill) with correct active-state highlighting.
- Adding a new server-rendered page follows a documented handler+template+route pattern.
