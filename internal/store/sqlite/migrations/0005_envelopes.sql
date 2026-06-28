-- 0005_envelopes.sql
-- Approval-envelope mirror (ADR 0054). The authoritative copy is the per-node
-- sidecar .groundwork/tickets/<id>/envelope.yaml (ADR 0053); this table is a
-- projection for live evaluation and queries. doc_json holds the full envelope.
CREATE TABLE envelopes (
  id          TEXT PRIMARY KEY,
  node_id     TEXT NOT NULL REFERENCES tickets(id) ON DELETE CASCADE,
  status      TEXT NOT NULL,            -- active | revoked | superseded
  approved_by TEXT,
  approved_at TEXT,
  doc_json    TEXT NOT NULL DEFAULT '{}',
  created_at  TEXT NOT NULL,
  updated_at  TEXT NOT NULL
);

CREATE INDEX idx_envelopes_node ON envelopes(node_id);

-- At most one active envelope per node (ADR 0054: one active envelope per
-- ancestor chain; per-node uniqueness is the storeable invariant).
CREATE UNIQUE INDEX idx_envelopes_active_node ON envelopes(node_id) WHERE status = 'active';
