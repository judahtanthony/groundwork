package sqlite

import (
	"database/sql"
	"encoding/json"

	"groundwork/internal/encoding"
)

// SetRunChangedFiles records a completed run's changed-file set (ADR 0059), the
// authoritative diff source for gate inputs. Order is normalized for stability.
func (db *DB) SetRunChangedFiles(runID string, files []string) error {
	if files == nil {
		files = []string{}
	}
	doc, err := encoding.JSON(files)
	if err != nil {
		return err
	}
	_, err = db.Exec(`UPDATE runs SET changed_files_json=?, updated_at=? WHERE id=?`,
		doc, encoding.Now(), runID)
	return err
}

// RunChangedFiles returns a run's recorded changed-file set.
func (db *DB) RunChangedFiles(runID string) ([]string, error) {
	var doc string
	err := db.QueryRow(`SELECT changed_files_json FROM runs WHERE id=?`, runID).Scan(&doc)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return decodeStringList(doc)
}

// ChangedFilesForNode returns the changed-file set of a node's most recent run
// that recorded one (ADR 0059). This is the diff the gate reads for envelope
// file-scope and escalation. Empty when no run produced a diff.
func (db *DB) ChangedFilesForNode(nodeID string) ([]string, error) {
	rows, err := db.Query(`SELECT changed_files_json FROM runs
		WHERE ticket_id=? ORDER BY started_at DESC, id DESC`, nodeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var doc string
		if err := rows.Scan(&doc); err != nil {
			return nil, err
		}
		files, err := decodeStringList(doc)
		if err != nil {
			return nil, err
		}
		if len(files) > 0 {
			return files, nil
		}
	}
	return nil, rows.Err()
}

// LatestRunIDForNode returns the most recent run id for a node, or "" when the
// node has no runs. Used by landing to locate the run's gw/run/<id> branch to
// squash (ADR 0059).
func (db *DB) LatestRunIDForNode(nodeID string) (string, error) {
	var id string
	err := db.QueryRow(`SELECT id FROM runs WHERE ticket_id=? ORDER BY started_at DESC, id DESC LIMIT 1`, nodeID).Scan(&id)
	if err == sql.ErrNoRows {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return id, nil
}

// LatestInterruptedRunForNode returns the most recent interrupted run id for a
// node (recovery.md), or "" — the run whose checkpoint chain a resume continues
// from (T-0904, ADR 0015).
func (db *DB) LatestInterruptedRunForNode(nodeID string) (string, error) {
	var id string
	err := db.QueryRow(`SELECT id FROM runs WHERE ticket_id=? AND status='interrupted'
		ORDER BY started_at DESC, id DESC LIMIT 1`, nodeID).Scan(&id)
	if err == sql.ErrNoRows {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return id, nil
}

func decodeStringList(doc string) ([]string, error) {
	if doc == "" {
		return nil, nil
	}
	var files []string
	if err := json.Unmarshal([]byte(doc), &files); err != nil {
		return nil, err
	}
	return files, nil
}
