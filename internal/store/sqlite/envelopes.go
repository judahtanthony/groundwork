package sqlite

import (
	"database/sql"
	"encoding/json"

	"groundwork/internal/encoding"
	"groundwork/internal/envelope"
)

// NextEnvelopeID allocates the next monotonic envelope id (ENV-0001, …).
func (db *DB) NextEnvelopeID() (string, error) {
	var id string
	err := db.withTx(func(tx *sql.Tx) error {
		var e error
		id, e = nextSeqID(tx, "envelope_seq", "ENV")
		return e
	})
	return id, err
}

// UpsertEnvelope mirrors an authoritative envelope (sidecar) into SQLite for
// live evaluation (ADR 0053). The sidecar remains the source of truth.
func (db *DB) UpsertEnvelope(e *envelope.Envelope) error {
	doc, err := json.Marshal(e)
	if err != nil {
		return err
	}
	now := encoding.Now()
	_, err = db.Exec(`INSERT INTO envelopes
		(id, node_id, status, approved_by, approved_at, doc_json, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			status=excluded.status, approved_by=excluded.approved_by,
			approved_at=excluded.approved_at, doc_json=excluded.doc_json,
			updated_at=excluded.updated_at`,
		e.ID, e.NodeID, string(e.Status), e.ApprovedBy, e.ApprovedAt, string(doc), now, now)
	return err
}

func envelopeFromDoc(doc string) (*envelope.Envelope, error) {
	var e envelope.Envelope
	if err := json.Unmarshal([]byte(doc), &e); err != nil {
		return nil, err
	}
	return &e, nil
}

// GetEnvelope returns an envelope by id, or ErrNotFound. The status column is
// authoritative for lifecycle state (SetEnvelopeStatus updates it), so it is
// overlaid onto the doc projection to avoid blob/column drift.
func (db *DB) GetEnvelope(id string) (*envelope.Envelope, error) {
	var doc, status string
	err := db.QueryRow(`SELECT doc_json, status FROM envelopes WHERE id = ?`, id).Scan(&doc, &status)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	e, err := envelopeFromDoc(doc)
	if err != nil {
		return nil, err
	}
	e.Status = envelope.Status(status)
	return e, nil
}

// GetActiveEnvelopeForNode returns the active envelope attached directly to
// nodeID, or (nil, nil) when the node has none.
func (db *DB) GetActiveEnvelopeForNode(nodeID string) (*envelope.Envelope, error) {
	var doc string
	err := db.QueryRow(`SELECT doc_json FROM envelopes WHERE node_id = ? AND status = 'active'`, nodeID).Scan(&doc)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return envelopeFromDoc(doc)
}

// ListEnvelopes returns all envelopes newest-first, optionally filtered by status.
func (db *DB) ListEnvelopes(status string) ([]*envelope.Envelope, error) {
	q := `SELECT doc_json, status FROM envelopes`
	var args []any
	if status != "" {
		q += ` WHERE status = ?`
		args = append(args, status)
	}
	q += ` ORDER BY id DESC`
	rows, err := db.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []*envelope.Envelope{}
	for rows.Next() {
		var doc, st string
		if err := rows.Scan(&doc, &st); err != nil {
			return nil, err
		}
		e, err := envelopeFromDoc(doc)
		if err != nil {
			return nil, err
		}
		e.Status = envelope.Status(st)
		out = append(out, e)
	}
	return out, rows.Err()
}

// SetEnvelopeStatus updates an envelope's lifecycle state (revoked/superseded)
// in the mirror; callers also rewrite the authoritative sidecar.
func (db *DB) SetEnvelopeStatus(id string, status envelope.Status) error {
	res, err := db.Exec(`UPDATE envelopes SET status = ?, updated_at = ? WHERE id = ?`,
		string(status), encoding.Now(), id)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return ErrNotFound
	}
	return nil
}
