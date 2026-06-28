-- 0006_integration_branches.sql
-- Per-root integration target (ADR 0058): the git branch a root's children land
-- to (land_to_parent) before the gated root land_to_main merge. Runtime state
-- (SQLite only); the branch itself lives in git.
CREATE TABLE integration_branches (
  node_id     TEXT PRIMARY KEY REFERENCES tickets(id) ON DELETE CASCADE,
  branch      TEXT NOT NULL,
  base_commit TEXT,
  status      TEXT NOT NULL,            -- open | landed
  created_at  TEXT NOT NULL,
  updated_at  TEXT NOT NULL
);
