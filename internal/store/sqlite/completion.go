package sqlite

import (
	"database/sql"
	"encoding/json"

	"groundwork/internal/completion"
	"groundwork/internal/encoding"
)

// UpsertCompletionSummary mirrors an authoritative completion sidecar into SQLite
// (ADR 0047/0057). The sidecar remains the source of truth.
func (db *DB) UpsertCompletionSummary(s *completion.Summary) error {
	doc, err := json.Marshal(s)
	if err != nil {
		return err
	}
	now := encoding.Now()
	_, err = db.Exec(`INSERT INTO completion_summaries (node_id, doc_json, created_at, updated_at)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(node_id) DO UPDATE SET doc_json=excluded.doc_json, updated_at=excluded.updated_at`,
		s.NodeID, string(doc), now, now)
	return err
}

// GetCompletionSummary returns a node's summary, or (nil, nil) when none exists.
func (db *DB) GetCompletionSummary(nodeID string) (*completion.Summary, error) {
	var doc string
	err := db.QueryRow(`SELECT doc_json FROM completion_summaries WHERE node_id = ?`, nodeID).Scan(&doc)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var s completion.Summary
	if err := json.Unmarshal([]byte(doc), &s); err != nil {
		return nil, err
	}
	return &s, nil
}
