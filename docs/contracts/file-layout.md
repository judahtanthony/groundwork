# File Layout Contract

Groundwork manages one dot directory in target repositories:

```text
.groundwork/
  config.yaml
  WORKFLOW.md
  state.sqlite
  tickets/
  policies/
  sops/
  runs/
  approvals/
  views/
  worktrees/
```

## Commit By Default

- `.groundwork/config.yaml`
- `.groundwork/WORKFLOW.md`
- `.groundwork/policies/*.yaml`
- `.groundwork/sops/<task-type>/**` (task-type SOPs and context)
- `.groundwork/tickets/**/ticket.md`
- `.groundwork/tickets/**/timeline.ndjson` when configured as durable audit export

## Ignore By Default

- `.groundwork/state.sqlite`
- `.groundwork/state.sqlite-wal`
- `.groundwork/state.sqlite-shm`
- `.groundwork/runs/`
- `.groundwork/approvals/`
- `.groundwork/views/`
- `.groundwork/worktrees/`

Generated views are never source of truth.

