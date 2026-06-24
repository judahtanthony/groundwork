# ADR 0021: Config Discovery And `config.yaml` Schema

Status: Accepted
Implemented: Implemented

## Context

`docs/contracts/cli.md` lets store-direct commands run without the coordinator, so they must locate `.groundwork/` themselves. `docs/contracts/file-layout.md` defines `config.yaml` as committed but does not specify its schema or the discovery algorithm.

## Decision

- **Discovery:** from the working directory, walk parent directories until a `.groundwork/` directory is found (mirroring git's `.git` search); its container is the project root. A `--root` flag and the `GW_ROOT` environment variable override the walk. Failure to find one is a clear, actionable error that suggests `gw init`.
- **Schema (`groundwork_config/v1`):** a minimal set of Phase 1 keys with defaults — schema version, `runtime: codex`, `server.addr: 127.0.0.1:4500`, `max_concurrency: 4`, `lease.ttl: 90s`, `lease.heartbeat: 30s`, and default sandbox posture `workspace-write`. Unknown keys produce a warning rather than an error (forward-safe).

## Consequences

One discovery path is shared by every store-direct command and, later, the server. `gw init` writes a minimal defaulted file. The versioned schema evolves cleanly as Phase 2 adds keys.
