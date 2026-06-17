package sqlite

import (
	"database/sql"

	"groundwork/internal/encoding"
)

// AuditEvent is one row of the append-only audit log.
type AuditEvent struct {
	ID         int64  `json:"id"`
	Actor      string `json:"actor"`
	Type       string `json:"type"`
	ObjectType string `json:"object_type"`
	ObjectID   string `json:"object_id"`
	Payload    string `json:"payload"`
	CreatedAt  string `json:"created_at"`
}

// appendAudit inserts an audit row within tx. payload is canonicalized to JSON;
// nil payload becomes "{}".
func appendAudit(tx *sql.Tx, actor, eventType, objectType, objectID string, payload any) error {
	payloadJSON := "{}"
	if payload != nil {
		j, err := encoding.JSON(payload)
		if err != nil {
			return err
		}
		payloadJSON = j
	}
	_, err := tx.Exec(
		`INSERT INTO audit_events (actor, type, object_type, object_id, payload_json, created_at)
		 VALUES (?,?,?,?,?,?)`,
		actor, eventType, objectType, objectID, payloadJSON, encoding.Now(),
	)
	return err
}

// AuditEventsFor returns audit events for an object, oldest first.
func (db *DB) AuditEventsFor(objectType, objectID string) ([]AuditEvent, error) {
	rows, err := db.Query(
		`SELECT id, actor, type, object_type, object_id, payload_json, created_at
		 FROM audit_events WHERE object_type=? AND object_id=? ORDER BY id`,
		objectType, objectID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []AuditEvent
	for rows.Next() {
		var e AuditEvent
		if err := rows.Scan(&e.ID, &e.Actor, &e.Type, &e.ObjectType, &e.ObjectID, &e.Payload, &e.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}
