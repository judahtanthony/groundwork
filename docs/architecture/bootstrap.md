# Bootstrap Architecture

This repository starts as documentation only because the product is intended to coordinate future agent work. The first durable artifact must therefore be the product and architecture memory that future agents will use.

Do not start by implementing code without this knowledge base. The risk is that early agents would bake in accidental assumptions about state, trust, persistence, and workflow.

## Bootstrap Phases

1. Record product, architecture, contracts, ADRs, and initial work tree.
2. Start a fresh implementation session from the docs.
3. Build CLI and SQLite store.
4. Build coordinator.
5. Import the bootstrap work tree into Groundwork.
6. Use Groundwork on low-risk docs and CLI tickets.
7. Add Codex runtime and dogfood real implementation work.

## Current Repo Status

Until implementation begins, this repo must not contain Go source, `go.mod`, generated assets, `.groundwork/`, or SQLite files.

