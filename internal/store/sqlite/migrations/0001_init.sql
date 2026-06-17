-- Groundwork initial schema (Phase 1).
-- Tables follow docs/contracts/sqlite-schema.md. Only the tables Phase 1
-- exercises are created here; runs/run_events/approvals/validation_results are
-- Phase 2 entities and are added by a later migration. Columns that Phase 2
-- needs on existing tables (parent_id, node_type, contract_json, risk_score)
-- are present now so the schema is forward-compatible (ADR 0012).

-- Key/value metadata, e.g. the ticket id sequence (ADR 0019).
CREATE TABLE meta (
  key   TEXT PRIMARY KEY,
  value TEXT NOT NULL
);

-- Work nodes. kind is advisory (ADR 0009); node_type is the structural fact,
-- set at triage (leaf | composite). reversibility is intentionally NOT a column
-- here: it is a per-action gate input evaluated at approval time (ADR 0014),
-- so it lives on the Phase 2 approvals row, not on the node.
CREATE TABLE tickets (
  id              TEXT PRIMARY KEY,
  parent_id       TEXT REFERENCES tickets(id) ON DELETE CASCADE,
  kind            TEXT NOT NULL DEFAULT 'ticket',
  node_type       TEXT,                                  -- leaf | composite
  title           TEXT NOT NULL,
  description     TEXT NOT NULL DEFAULT '',
  contract_json   TEXT NOT NULL DEFAULT '{}',            -- parent contract for composites
  status          TEXT NOT NULL,
  assignee        TEXT,
  priority        INTEGER,
  labels_json     TEXT NOT NULL DEFAULT '[]',
  acceptance_json TEXT NOT NULL DEFAULT '[]',
  risk_score      INTEGER,                               -- 0-100, display/ranking
  created_at      TEXT NOT NULL,
  updated_at      TEXT NOT NULL
);

CREATE INDEX idx_tickets_parent ON tickets(parent_id);
CREATE INDEX idx_tickets_status ON tickets(status);

-- Dependency edges: a directed-acyclic overlay (ADR 0010). from_id depends on
-- to_id; cycles are rejected in application code (T-0113).
CREATE TABLE dependencies (
  from_id    TEXT NOT NULL REFERENCES tickets(id) ON DELETE CASCADE,
  to_id      TEXT NOT NULL REFERENCES tickets(id) ON DELETE CASCADE,
  created_at TEXT NOT NULL,
  PRIMARY KEY (from_id, to_id)
);

CREATE INDEX idx_dependencies_to ON dependencies(to_id);

-- Exclusive active-work claims (T-0117). run_id is a plain TEXT in Phase 1
-- (the runs table is Phase 2); the claim primitive is exercised on its own.
CREATE TABLE leases (
  ticket_id  TEXT PRIMARY KEY REFERENCES tickets(id) ON DELETE CASCADE,
  run_id     TEXT NOT NULL,
  agent_id   TEXT NOT NULL,
  status     TEXT NOT NULL,
  expires_at TEXT NOT NULL,
  renewed_at TEXT NOT NULL
);

-- Append-only audit log. Every state-changing operation appends one row in the
-- same transaction as the change (docs/contracts/sqlite-schema.md).
CREATE TABLE audit_events (
  id           INTEGER PRIMARY KEY AUTOINCREMENT,
  actor        TEXT NOT NULL,
  type         TEXT NOT NULL,
  object_type  TEXT NOT NULL,
  object_id    TEXT NOT NULL,
  payload_json TEXT NOT NULL DEFAULT '{}',
  created_at   TEXT NOT NULL
);

CREATE INDEX idx_audit_object ON audit_events(object_type, object_id);
