-- 0004_policy_suggestions.sql
-- Elevation-readiness suggestion queue (ADR 0038). The system proposes autonomy
-- elevations for human review; it never self-elevates. Promoting a suggestion
-- emits the policy change for a human to apply — amend_policy / elevate_autonomy
-- remain human-gated actions, so this table is advisory only.
CREATE TABLE policy_suggestions (
  id          TEXT PRIMARY KEY,
  kind        TEXT NOT NULL,            -- elevate_autonomy
  action_type TEXT NOT NULL,            -- execute | decompose | ...
  work_type   TEXT NOT NULL,
  level       TEXT NOT NULL,            -- suggested autonomy level (e.g. auto)
  rationale   TEXT NOT NULL,
  status      TEXT NOT NULL,            -- pending | promoted | dismissed
  created_at  TEXT NOT NULL,
  decided_at  TEXT
);

CREATE INDEX idx_policy_suggestions_status ON policy_suggestions(status);

-- At most one pending suggestion per elevation target.
CREATE UNIQUE INDEX idx_policy_suggestions_pending
  ON policy_suggestions(kind, action_type, work_type) WHERE status = 'pending';
