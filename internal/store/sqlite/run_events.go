package sqlite

import (
	"database/sql"

	"groundwork/internal/encoding"
)

// RunEvent is one row of append-only run telemetry.
type RunEvent struct {
	ID        int64  `json:"id"`
	RunID     string `json:"run_id"`
	EventType string `json:"event_type"`
	Payload   string `json:"payload"`
	CreatedAt string `json:"created_at"`
}

// AppendRunEvent records a run event and refreshes the run's last_event /
// last_message / updated_at so run listings show current activity. payload is
// canonicalized JSON; a nil payload becomes "{}".
func (db *DB) AppendRunEvent(runID, eventType, message string, payload any) (*RunEvent, error) {
	payloadJSON := "{}"
	if payload != nil {
		j, err := encoding.JSON(payload)
		if err != nil {
			return nil, err
		}
		payloadJSON = j
	}
	now := encoding.Now()
	var ev RunEvent
	err := db.withTx(func(tx *sql.Tx) error {
		res, err := tx.Exec(
			`INSERT INTO run_events (run_id, event_type, payload_json, created_at) VALUES (?,?,?,?)`,
			runID, eventType, payloadJSON, now,
		)
		if err != nil {
			return err
		}
		id, _ := res.LastInsertId()
		if message != "" {
			_, err = tx.Exec(`UPDATE runs SET last_event=?, last_message=?, updated_at=? WHERE id=?`,
				eventType, message, now, runID)
		} else {
			_, err = tx.Exec(`UPDATE runs SET last_event=?, updated_at=? WHERE id=?`,
				eventType, now, runID)
		}
		if err != nil {
			return err
		}
		ev = RunEvent{ID: id, RunID: runID, EventType: eventType, Payload: payloadJSON, CreatedAt: now}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &ev, nil
}

// ListRunEvents returns a run's events, oldest first.
func (db *DB) ListRunEvents(runID string) ([]RunEvent, error) {
	rows, err := db.Query(
		`SELECT id, run_id, event_type, payload_json, created_at FROM run_events WHERE run_id=? ORDER BY id`,
		runID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []RunEvent{}
	for rows.Next() {
		var e RunEvent
		if err := rows.Scan(&e.ID, &e.RunID, &e.EventType, &e.Payload, &e.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

// RecordCheckpoint records a run's work-in-progress checkpoint (ADR 0015). In M2
// this is a record only: the actual WIP commit on the worktree branch and the
// refs/groundwork/runs/<run-id> namespace are the Phase 4 runtime's job
// (ADR 0027). ref is the (future) checkpoint commit/ref identifier.
func (db *DB) RecordCheckpoint(runID, ref string) (*RunEvent, error) {
	return db.AppendRunEvent(runID, "checkpoint", "", map[string]any{"ref": ref})
}

// SquashCheckpoints records that a run's WIP checkpoints were squashed at
// landing (ADR 0015): main sees one clean commit and the WIP chain is dropped.
// The git squash itself is Phase 4; M2 records the lifecycle event.
func (db *DB) SquashCheckpoints(runID string) (*RunEvent, error) {
	return db.AppendRunEvent(runID, "checkpoints_squashed", "", nil)
}
