---
name: gw
description: Drive the Groundwork `gw` CLI on the user's behalf — inspect and act on the work tree (tickets, status/board, approvals, runs, validation, landing). Use when the user asks about their Groundwork work: "what needs my approval", "what's on the board", "claim the next ticket", "show ticket G-0007", "land this", "what's running", or any request to read or change Groundwork state.
---

# Driving the `gw` CLI

`gw` is the Groundwork CLI. It administers this project's work tree (tickets), runs,
approvals, and landing gate. You run it through the Bash tool on the user's behalf.

This skill is intentionally thin: the CLI is **self-documenting**, so discover the live
surface at runtime rather than trusting a frozen list here.

## Discover the live surface (do this first if unsure)

- `gw help` — the full command tree.
- `gw <command>` — a group's subcommands (e.g. `gw ticket`, `gw approval`).
- `gw <command> <subcommand> -h` — usage for one command (e.g. `gw ticket transition -h`).

As the CLI grows, these outputs are the source of truth. Prefer them over guessing.

## Always prefer JSON

Pass `--json` (trailing) on any command you need to parse:

```
gw status --json
gw ticket list --json
gw approval list --json
```

Parse the JSON; don't scrape the human table format.

## Command groups (current)

| Group | What it does |
|---|---|
| `gw status` / `gw board` | Work-tree summary; tickets grouped by status |
| `gw ticket` | create, list, show, edit, transition, triage, tree, link, context, decompose, escalate, land, export, import |
| `gw approval` | list, show, approve, reject, clarify |
| `gw run` | inspect and control runs |
| `gw validation` | list and run validation checks |
| `gw context` | bounded context brief for a node |
| `gw actor` | inspect the local actor registry |
| `gw server` | run the localhost coordinator (HTTP API + SSE) |
| `gw init` / `gw export` / `gw doctor` / `gw version` | setup, export, diagnostics, version |

## Coordinator-backed commands

`approval`, `run`, `decompose`, `escalate`, and `land` require the coordinator. If a
command returns `{"error":{"code":"coordinator_required"}}`, the server isn't running —
tell the user and offer to start it with `gw server` (long-running; run in background).
Do not silently work around it.

## Actors

The human is `human.owner`; the AI implementer is `ai.codex.default` (`gw actor list`).
When the user says "me" / "my", they mean `human.owner`.

## Common request → command

- **"What needs my approval?"** → `gw approval list --json` (needs coordinator), then
  `gw approval show <id>` for details.
- **"What's on the board / status?"** → `gw board --json` or `gw status --json`.
- **"Claim the next ticket available to me."** → there is no single `claim` command.
  Find an eligible leaf (status `todo`, dependencies satisfied, allowed for the actor)
  via `gw board --json` / `gw ticket tree --json`, confirm the choice with the user,
  then `gw ticket transition <id> in_progress`. Check valid statuses with
  `gw ticket transition -h`.
- **"Show ticket X."** → `gw ticket show <id> --json`.
- **"Add a ticket."** → `gw ticket create -h` for flags, then create it.

## Boundary — act within the v1 trust model (see AGENTS.md)

Groundwork keeps humans in the loop. When acting as an agent:

- **Reads are free.** Run any inspecting command (`status`, `board`, `list`, `show`,
  `tree`, `context`) without asking.
- **State changes need the user's intent.** Only transition/create/edit/link tickets,
  approve/reject, or land when the user actually asked for that change. Confirm the
  specific target first when it's ambiguous (e.g. which ticket to claim).
- **Landing and decomposition are human-gated.** `gw ticket land` and
  `gw ticket decompose` go through approval gates. Never approve on the human's behalf
  to push something through, and never self-elevate past a gate.
- **Don't loosen gates or policy** to get around a block. Surface the block instead.
