-- 0007_completion_summaries.sql
-- Child completion-summary mirror (ADR 0047/0057). Authoritative copy is the
-- per-node sidecar .groundwork/tickets/<id>/completion.yaml; this projection lets
-- the bulk review bundle aggregate summaries by query. doc_json holds the record.
CREATE TABLE completion_summaries (
  node_id    TEXT PRIMARY KEY REFERENCES tickets(id) ON DELETE CASCADE,
  doc_json   TEXT NOT NULL DEFAULT '{}',
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
);
