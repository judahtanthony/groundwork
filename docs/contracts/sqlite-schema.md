# SQLite Schema Contract

SQLite is the v1 operational store. Use WAL mode and foreign keys.

Recommended pragmas:

```sql
PRAGMA journal_mode = WAL;
PRAGMA foreign_keys = ON;
PRAGMA busy_timeout = 5000;
```

Draft minimum schema:

```sql
-- tickets are work nodes; kind is advisory, node_type is structural.
tickets(
  id TEXT PRIMARY KEY,
  parent_id TEXT REFERENCES tickets(id) ON DELETE CASCADE,
  kind TEXT NOT NULL DEFAULT 'ticket',
  node_type TEXT,                 -- leaf | composite, set at triage
  title TEXT NOT NULL,
  description TEXT NOT NULL DEFAULT '',
  contract_json TEXT NOT NULL DEFAULT '{}',  -- parent design/contract for composite nodes
  status TEXT NOT NULL,
  assignee TEXT,
  priority INTEGER,
  labels_json TEXT NOT NULL DEFAULT '[]',
  acceptance_json TEXT NOT NULL DEFAULT '[]',
  risk_score INTEGER,             -- last computed 0–100 score (display/ranking)
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
);

-- dependency edges form a DAG overlay; cycles are rejected.
dependencies(
  from_id TEXT NOT NULL REFERENCES tickets(id) ON DELETE CASCADE,
  to_id TEXT NOT NULL REFERENCES tickets(id) ON DELETE CASCADE,
  created_at TEXT NOT NULL,
  PRIMARY KEY (from_id, to_id)
);

leases(
  ticket_id TEXT PRIMARY KEY REFERENCES tickets(id) ON DELETE CASCADE,
  run_id TEXT NOT NULL,
  agent_id TEXT NOT NULL,
  status TEXT NOT NULL,
  expires_at TEXT NOT NULL,
  renewed_at TEXT NOT NULL
);

runs(
  id TEXT PRIMARY KEY,
  ticket_id TEXT NOT NULL REFERENCES tickets(id) ON DELETE CASCADE,
  agent_id TEXT NOT NULL,
  status TEXT NOT NULL,
  workspace_path TEXT NOT NULL,
  base_commit TEXT,
  started_at TEXT NOT NULL,
  updated_at TEXT NOT NULL,
  completed_at TEXT,
  last_event TEXT,
  last_message TEXT,
  input_tokens INTEGER NOT NULL DEFAULT 0,
  output_tokens INTEGER NOT NULL DEFAULT 0,
  total_tokens INTEGER NOT NULL DEFAULT 0
);

run_events(
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  run_id TEXT NOT NULL REFERENCES runs(id) ON DELETE CASCADE,
  event_type TEXT NOT NULL,
  payload_json TEXT NOT NULL,
  created_at TEXT NOT NULL
);

approvals(
  id TEXT PRIMARY KEY,
  run_id TEXT NOT NULL REFERENCES runs(id) ON DELETE CASCADE,
  ticket_id TEXT NOT NULL REFERENCES tickets(id) ON DELETE CASCADE,
  type TEXT NOT NULL,             -- e.g. execute | land_to_main | decompose | replan
  risk_class TEXT NOT NULL,       -- low | medium | high | critical
  risk_score INTEGER,             -- 0–100; class is what gates key off
  summary TEXT NOT NULL,
  action_json TEXT NOT NULL,
  status TEXT NOT NULL,
  requested_by TEXT NOT NULL,
  decided_by TEXT,
  decision_reason TEXT,
  created_at TEXT NOT NULL,
  decided_at TEXT
);

validation_results(
  id TEXT PRIMARY KEY,
  ticket_id TEXT NOT NULL REFERENCES tickets(id) ON DELETE CASCADE,
  run_id TEXT REFERENCES runs(id) ON DELETE SET NULL,
  name TEXT NOT NULL,
  command TEXT,
  status TEXT NOT NULL,
  artifact_path TEXT,
  started_at TEXT,
  completed_at TEXT
);

audit_events(
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  actor TEXT NOT NULL,
  type TEXT NOT NULL,
  object_type TEXT NOT NULL,
  object_id TEXT NOT NULL,
  payload_json TEXT NOT NULL,
  created_at TEXT NOT NULL
);
```

Decomposition proposals (`decompose`) and re-plan decisions (`replan`) reuse `approvals` with the corresponding `type`; escalation / upward-revision events are recorded in `audit_events` (and the node timeline). A decomposition creates child nodes in `backlog` (non-dispatchable) until the proposal is approved, after which they move to `todo` as dependencies allow. SOPs and task-type context are file-based under `.groundwork/sops/`, not database rows; autonomy levels live in the policy YAML.

All state-changing operations must run in transactions and append an audit event.

