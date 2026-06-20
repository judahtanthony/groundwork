package sqlite

import (
	"database/sql"

	"groundwork/internal/encoding"
)

// Validation result statuses.
const (
	ValidationPass    = "pass"
	ValidationFail    = "fail"
	ValidationRunning = "running"
	ValidationSkipped = "skipped"
)

// ValidationResult is one recorded validation outcome (sqlite-schema.md).
type ValidationResult struct {
	ID           string `json:"id"`
	TicketID     string `json:"ticket_id"`
	RunID        string `json:"run_id,omitempty"`
	Name         string `json:"name"`
	Command      string `json:"command,omitempty"`
	Status       string `json:"status"`
	ArtifactPath string `json:"artifact_path,omitempty"`
	StartedAt    string `json:"started_at,omitempty"`
	CompletedAt  string `json:"completed_at,omitempty"`
}

// RecordValidation stores a validation result linked to a ticket (and optionally
// a run) and appends an audit event.
func (db *DB) RecordValidation(v ValidationResult) (*ValidationResult, error) {
	now := encoding.Now()
	if v.CompletedAt == "" && (v.Status == ValidationPass || v.Status == ValidationFail) {
		v.CompletedAt = now
	}
	err := db.withTx(func(tx *sql.Tx) error {
		id, err := nextSeqID(tx, "validation_seq", "V")
		if err != nil {
			return err
		}
		v.ID = id
		_, err = tx.Exec(`INSERT INTO validation_results
			(id, ticket_id, run_id, name, command, status, artifact_path, started_at, completed_at)
			VALUES (?,?,?,?,?,?,?,?,?)`,
			v.ID, v.TicketID, nullStr(v.RunID), v.Name, nullStr(v.Command), v.Status,
			nullStr(v.ArtifactPath), nullStr(v.StartedAt), nullStr(v.CompletedAt))
		if err != nil {
			return err
		}
		return appendAudit(tx, "validation", "validation.recorded", "ticket", v.TicketID, map[string]any{
			"name": v.Name, "status": v.Status,
		})
	})
	if err != nil {
		return nil, err
	}
	return &v, nil
}

// ListValidationsForTicket returns a ticket's validation results, oldest first.
func (db *DB) ListValidationsForTicket(ticketID string) ([]*ValidationResult, error) {
	rows, err := db.Query(`SELECT id, ticket_id, run_id, name, command, status, artifact_path, started_at, completed_at
		FROM validation_results WHERE ticket_id=? ORDER BY id`, ticketID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []*ValidationResult{}
	for rows.Next() {
		var (
			v        ValidationResult
			runID    sql.NullString
			command  sql.NullString
			artifact sql.NullString
			started  sql.NullString
			done     sql.NullString
		)
		if err := rows.Scan(&v.ID, &v.TicketID, &runID, &v.Name, &command, &v.Status, &artifact, &started, &done); err != nil {
			return nil, err
		}
		v.RunID = runID.String
		v.Command = command.String
		v.ArtifactPath = artifact.String
		v.StartedAt = started.String
		v.CompletedAt = done.String
		out = append(out, &v)
	}
	return out, rows.Err()
}
