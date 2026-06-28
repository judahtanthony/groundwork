package sqlite

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"groundwork/internal/decision"
	"groundwork/internal/encoding"
)

// AppendDecision records a new durable decision event for a ticket (ADR 0051).
// When rec.Sequence is 0 the next per-ticket sequence is assigned; otherwise the
// caller's sequence is preserved (import path). The sidecar remains the source of
// truth (ADR 0053); this is the live projection. Returns the stored record with
// its assigned sequence.
func (db *DB) AppendDecision(rec decision.Record) (decision.Record, error) {
	if err := rec.Validate(); err != nil {
		return decision.Record{}, err
	}
	err := db.withTx(func(tx *sql.Tx) error {
		if rec.Sequence == 0 {
			var max sql.NullInt64
			if err := tx.QueryRow(`SELECT MAX(seq) FROM decisions WHERE ticket_id = ?`, rec.TicketID).Scan(&max); err != nil {
				return err
			}
			rec.Sequence = int(max.Int64) + 1
		}
		return insertDecision(tx, rec)
	})
	if err != nil {
		return decision.Record{}, err
	}
	return rec, nil
}

// ImportDecision inserts a decision record verbatim from a sidecar, preserving
// its sequence and id so a rebuilt store re-exports byte-for-byte. It upserts on
// (ticket_id, seq) so re-import is idempotent.
func (db *DB) ImportDecision(rec decision.Record) error {
	if err := rec.Validate(); err != nil {
		return err
	}
	if rec.Sequence == 0 {
		return fmt.Errorf("import decision %q: sequence is required", rec.TicketID)
	}
	return db.withTx(func(tx *sql.Tx) error { return insertDecision(tx, rec) })
}

func insertDecision(tx *sql.Tx, rec decision.Record) error {
	doc, err := json.Marshal(&rec)
	if err != nil {
		return err
	}
	created := rec.RequestedAt
	if created == "" {
		created = rec.DecidedAt
	}
	if created == "" {
		created = encoding.Now()
	}
	_, err = tx.Exec(`INSERT INTO decisions (ticket_id, seq, decision_id, event_type, status, doc_json, created_at)
		VALUES (?,?,?,?,?,?,?)
		ON CONFLICT(ticket_id, seq) DO UPDATE SET
			decision_id=excluded.decision_id, event_type=excluded.event_type,
			status=excluded.status, doc_json=excluded.doc_json, created_at=excluded.created_at`,
		rec.TicketID, rec.Sequence, rec.ID, rec.EventType, rec.Status, string(doc), created)
	return err
}

// ListDecisions returns a ticket's decision records in append (sequence) order.
func (db *DB) ListDecisions(ticketID string) ([]decision.Record, error) {
	rows, err := db.Query(`SELECT doc_json FROM decisions WHERE ticket_id = ? ORDER BY seq`, ticketID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []decision.Record
	for rows.Next() {
		var doc string
		if err := rows.Scan(&doc); err != nil {
			return nil, err
		}
		var r decision.Record
		if err := json.Unmarshal([]byte(doc), &r); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

// ListPendingDecisions returns all decision records currently in pending status,
// across tickets — the set the live queues project from on rebuild (ADR 0051).
func (db *DB) ListPendingDecisions() ([]decision.Record, error) {
	rows, err := db.Query(`SELECT doc_json FROM decisions WHERE status = ? ORDER BY ticket_id, seq`, decision.StatusPending)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []decision.Record
	for rows.Next() {
		var doc string
		if err := rows.Scan(&doc); err != nil {
			return nil, err
		}
		var r decision.Record
		if err := json.Unmarshal([]byte(doc), &r); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

// HasDecisions reports whether a ticket has any durable decision records.
func (db *DB) HasDecisions(ticketID string) (bool, error) {
	var n int
	if err := db.QueryRow(`SELECT COUNT(*) FROM decisions WHERE ticket_id = ?`, ticketID).Scan(&n); err != nil {
		return false, err
	}
	return n > 0, nil
}
