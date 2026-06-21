package sqlite

import (
	"database/sql"
	"fmt"

	"groundwork/internal/encoding"
	"groundwork/internal/ticket"
)

// ImportTicket inserts a node from a committed export, preserving its id and
// timestamps (recovery.md, T-0902). Unlike CreateTicket it does not allocate an
// id or stamp "now", so a cold store rebuilt from exports matches the originals.
// The parent (if any) must already exist (foreign key). It appends a
// ticket.imported audit event. Runtime state (leases, runs) is never restored.
func (db *DB) ImportTicket(t *ticket.Ticket) error {
	if t.ID == "" {
		return fmt.Errorf("import requires a ticket id")
	}
	if err := prepareTicket(t); err != nil {
		return err
	}
	if t.CreatedAt == "" {
		t.CreatedAt = encoding.Now()
	}
	if t.UpdatedAt == "" {
		t.UpdatedAt = t.CreatedAt
	}
	return db.withTx(func(tx *sql.Tx) error {
		labels, acceptance, err := marshalLists(t)
		if err != nil {
			return err
		}
		_, err = tx.Exec(`INSERT INTO tickets
			(id, parent_id, kind, node_type, work_type, title, description, contract_json,
			 status, assignee, requested_actor, priority, labels_json, acceptance_json,
			 risk_score, created_at, updated_at)
			VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
			t.ID, nullStr(t.ParentID), t.Kind, nullStr(string(t.NodeType)), nullStr(t.WorkType),
			t.Title, t.Description, t.Contract, string(t.Status), nullStr(t.Assignee),
			nullStr(t.RequestedActor), nullFloat(t.Priority), labels, acceptance,
			nullInt(t.RiskScore), t.CreatedAt, t.UpdatedAt)
		if err != nil {
			return err
		}
		return appendAudit(tx, "import", "ticket.imported", "ticket", t.ID, map[string]any{
			"title": t.Title,
		})
	})
}

// HasTickets reports whether any ticket rows exist (used to decide cold-start
// import).
func (db *DB) HasTickets() (bool, error) {
	var n int
	if err := db.QueryRow(`SELECT COUNT(*) FROM tickets`).Scan(&n); err != nil {
		return false, err
	}
	return n > 0, nil
}
