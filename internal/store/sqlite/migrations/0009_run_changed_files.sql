-- 0009_run_changed_files.sql
-- The run's changed-file set captured from its isolated worktree (ADR 0059): the
-- authoritative diff for gate inputs (validation template selection, envelope
-- file-scope, escalation triggers). The full unified diff is run evidence under
-- .groundwork/runs/<id>/diff.patch (ignored); this column is the queryable list.
ALTER TABLE runs ADD COLUMN changed_files_json TEXT NOT NULL DEFAULT '[]';
