# Coordinator

The coordinator is the local process started by `gw server`.

## Responsibilities

- Own multi-agent scheduling decisions.
- Open and manage SQLite access.
- Claim eligible nodes transactionally, where eligibility requires all dependencies satisfied.
- Match eligible nodes to actors using work type, requested actor/capability hints, risk, file scope, and policy.
- Create and supervise runs, including planning (decomposition) and implementation runs.
- Record actor identity and actor configuration snapshots on runs.
- Renew and expire leases.
- Pause, resume, and cancel runs.
- Route approval requests, including `decompose` proposals and escalation / re-plan decisions.
- Enforce actor-aware trust policy and validation gates.
- Export ticket, run, and approval projections.
- Expose dashboard, HTTP API, and SSE stream.

## Active Runs Require Coordinator

SQLite supports multi-process access, but active agent orchestration should go through the coordinator. This avoids split-brain scheduling, inconsistent approvals, and duplicate run lifecycle logic.

Simple ticket/config CLI operations may open SQLite directly if the coordinator is not running, using the same transaction library. Live run control should fail clearly without the coordinator.
