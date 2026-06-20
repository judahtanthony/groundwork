-- Phase 2 coordinator entities (T-0410). Adds the runtime/approval/validation
-- tables that docs/contracts/sqlite-schema.md defines but Phase 1 deferred.
-- Forward-only and additive (ADR 0018).

-- Runs are supervised node attempts. mode (planning | implementation) is a run
-- record field per ADR 0027 / runtime-model.md. actor_snapshot_json preserves
-- the selected actor configuration for audit after actors.yaml changes (ADR 0023).
CREATE TABLE runs (
  id                  TEXT PRIMARY KEY,
  ticket_id           TEXT NOT NULL REFERENCES tickets(id) ON DELETE CASCADE,
  actor_id            TEXT NOT NULL,
  actor_snapshot_json TEXT NOT NULL DEFAULT '{}',
  mode                TEXT NOT NULL,                 -- planning | implementation (ADR 0027)
  runtime             TEXT NOT NULL,                 -- e.g. codex
  model               TEXT,
  status              TEXT NOT NULL,
  workspace_path      TEXT NOT NULL DEFAULT '',
  base_commit         TEXT,
  started_at          TEXT NOT NULL,
  updated_at          TEXT NOT NULL,
  completed_at        TEXT,
  last_event          TEXT,
  last_message        TEXT,
  input_tokens        INTEGER NOT NULL DEFAULT 0,
  output_tokens       INTEGER NOT NULL DEFAULT 0,
  total_tokens        INTEGER NOT NULL DEFAULT 0
);

CREATE INDEX idx_runs_ticket ON runs(ticket_id);
CREATE INDEX idx_runs_status ON runs(status);

-- Append-only run telemetry; the SSE hub and JSONL logs project from here.
CREATE TABLE run_events (
  id           INTEGER PRIMARY KEY AUTOINCREMENT,
  run_id       TEXT NOT NULL REFERENCES runs(id) ON DELETE CASCADE,
  event_type   TEXT NOT NULL,
  payload_json TEXT NOT NULL DEFAULT '{}',
  created_at   TEXT NOT NULL
);

CREATE INDEX idx_run_events_run ON run_events(run_id, id);

-- Capability gates. decompose proposals and replan decisions reuse this table
-- via type. reversible records the per-action reversibility verdict that forces
-- critical when false (ADR 0014); risk_class is what gates key off, score ranks.
CREATE TABLE approvals (
  id                   TEXT PRIMARY KEY,
  run_id               TEXT REFERENCES runs(id) ON DELETE SET NULL,  -- null for human/system-initiated gates
  ticket_id            TEXT NOT NULL REFERENCES tickets(id) ON DELETE CASCADE,
  type                 TEXT NOT NULL,                -- execute | land_to_main | decompose | replan
  risk_class           TEXT NOT NULL,               -- low | medium | high | critical
  risk_score           INTEGER,                     -- 0-100, display/ranking
  reversible           INTEGER,                     -- 1 reversible, 0 irreversible (ADR 0014)
  summary              TEXT NOT NULL,
  action_json          TEXT NOT NULL DEFAULT '{}',
  status               TEXT NOT NULL,
  requested_by_actor   TEXT NOT NULL,
  decided_by_actor     TEXT,
  required_actors_json TEXT NOT NULL DEFAULT '[]',
  required_roles_json  TEXT NOT NULL DEFAULT '[]',
  decision_reason      TEXT,
  created_at           TEXT NOT NULL,
  decided_at           TEXT
);

CREATE INDEX idx_approvals_ticket ON approvals(ticket_id);
CREATE INDEX idx_approvals_status ON approvals(status);

-- Validation outcomes linked to a ticket and (optionally) the run that produced
-- them; artifact paths point into the ignored run log tree.
CREATE TABLE validation_results (
  id            TEXT PRIMARY KEY,
  ticket_id     TEXT NOT NULL REFERENCES tickets(id) ON DELETE CASCADE,
  run_id        TEXT REFERENCES runs(id) ON DELETE SET NULL,
  name          TEXT NOT NULL,
  command       TEXT,
  status        TEXT NOT NULL,
  artifact_path TEXT,
  started_at    TEXT,
  completed_at  TEXT
);

CREATE INDEX idx_validation_ticket ON validation_results(ticket_id);
