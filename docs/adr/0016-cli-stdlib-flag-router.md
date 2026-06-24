# ADR 0016: Build The CLI On Stdlib `flag` With A Minimal Subcommand Router

Status: Accepted
Implemented: Implemented

## Context

`docs/contracts/cli.md` defines a large noun-verb command tree (roughly fifty commands), `--json` on every data command, and a consistent JSON error envelope. `docs/reference/conventions.md` prefers the Go standard library "when clean enough" and otherwise small, mature, open-source dependencies. The realistic alternatives are `cobra` (ubiquitous but heavy, and a gateway to `viper`) and small `flag`-based libraries such as `peterbourgon/ff/v3/ffcli`.

## Decision

Build the CLI on the standard-library `flag` package plus a small in-tree subcommand router. Each command is a value carrying a name, its own `flag.FlagSet`, and a run function. Command handlers are framework-agnostic (plain functions over a store/client), so the router can be swapped without touching command logic. Global flags (`--json`, `--root`) and the error envelope are handled centrally.

## Consequences

No third-party CLI dependency in v1; full control over help, output, and JSON. We own subcommand help generation. If the command tree becomes unwieldy in a later phase, `peterbourgon/ff/v3/ffcli` (also `flag`-based, tiny) is the pre-approved fallback, and the framework-agnostic handlers keep that swap contained.
