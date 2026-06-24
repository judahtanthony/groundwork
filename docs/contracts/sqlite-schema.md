# SQLite Schema Contract

SQLite is the v1 live projection and runtime coordination store. Durable project state
is filesystem-authoritative and must be rebuildable from committed/exported files plus
git (ADR 0053). Use WAL mode and foreign keys.

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
  work_type TEXT,                 -- organization-defined operational type for SOP/policy/routing
  title TEXT NOT NULL,
  description TEXT NOT NULL DEFAULT '',
  contract_json TEXT NOT NULL DEFAULT '{}',  -- parent design/contract for composite nodes
  status TEXT NOT NULL,
  assignee TEXT,                  -- human-readable ownership label (display only)
  requested_actor TEXT,           -- optional actor routing hint, still policy-checked
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
  actor_id TEXT NOT NULL,
  status TEXT NOT NULL,
  expires_at TEXT NOT NULL,
  renewed_at TEXT NOT NULL
);

runs(
  id TEXT PRIMARY KEY,
  ticket_id TEXT NOT NULL REFERENCES tickets(id) ON DELETE CASCADE,
  actor_id TEXT NOT NULL,         -- actor selected from .groundwork/actors.yaml
  actor_snapshot_json TEXT NOT NULL DEFAULT '{}',
  mode TEXT NOT NULL,             -- planning | implementation (ADR 0027)
  runtime TEXT NOT NULL,          -- e.g. codex
  model TEXT,
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
  durable_request_id TEXT,        -- stable id from ticket decisions.ndjson when durable
  run_id TEXT REFERENCES runs(id) ON DELETE SET NULL,  -- null for human/system-initiated gates (decompose, replan)
  ticket_id TEXT NOT NULL REFERENCES tickets(id) ON DELETE CASCADE,
  type TEXT NOT NULL,             -- e.g. execute | land_to_main | decompose | replan
  risk_class TEXT NOT NULL,       -- low | medium | high | critical
  risk_score INTEGER,             -- 0–100; class is what gates key off
  reversible INTEGER,             -- 1 reversible, 0 irreversible; false forces critical (ADR 0014)
  summary TEXT NOT NULL,
  action_json TEXT NOT NULL,
  status TEXT NOT NULL,
  requested_by_actor TEXT NOT NULL,
  decided_by_actor TEXT,
  required_actors_json TEXT NOT NULL DEFAULT '[]',
  required_roles_json TEXT NOT NULL DEFAULT '[]',
  decision_reason TEXT,
  created_at TEXT NOT NULL,
  decided_at TEXT
);

decision_records(
  id TEXT NOT NULL,               -- stable durable request/decision id
  sequence INTEGER NOT NULL,
  ticket_id TEXT NOT NULL REFERENCES tickets(id) ON DELETE CASCADE,
  run_id TEXT REFERENCES runs(id) ON DELETE SET NULL,
  event_type TEXT NOT NULL,       -- decision_requested | input_requested | approval_requested | ...
  request_type TEXT,
  work_type TEXT,
  status TEXT NOT NULL,           -- pending | answered | accepted | rejected | superseded | recovered
  requested_by_actor TEXT,
  requested_at TEXT,
  decided_by_actor TEXT,
  decided_at TEXT,
  payload_json TEXT NOT NULL,
  created_at TEXT NOT NULL,
  PRIMARY KEY (ticket_id, id, sequence)
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

Actor definitions are file-based under `.groundwork/actors.yaml`, not database rows. Runs store an actor snapshot because actor files may change after a run completes. Decomposition proposals (`decompose`) and re-plan decisions (`replan`) create durable decision records when their payload must survive rebuild; `approvals` is then the live coordinator queue/projection over those records. Escalation / upward-revision events are recorded in `audit_events` and the node timeline, with durable blockers mirrored in `decision_records`. A decomposition creates child nodes in `backlog` (non-dispatchable) until the proposal is approved, after which they move to `todo` as dependencies allow. SOPs and work-type context are file-based under `.groundwork/sops/`, not database rows; autonomy levels live in the policy YAML.

All state-changing operations must run in transactions and append an audit event. If the
operation mutates durable project state, it must also update the filesystem source of
truth or an explicit durable replay record before reporting durable success.
