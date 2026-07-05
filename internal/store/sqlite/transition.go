package sqlite

import (
	"database/sql"
	"errors"
	"fmt"

	"groundwork/internal/encoding"
	"groundwork/internal/ticket"
)

// ErrIllegalTransition is returned when a status transition is not permitted by
// the Phase 1 transition map (ADR 0022).
var ErrIllegalTransition = errors.New("illegal status transition")

// TransitionTicket changes a ticket's status, validating the move against the
// Phase 1 transition map and appending a ticket.transitioned audit event with
// the from/to states.
func (db *DB) TransitionTicket(id string, to ticket.Status, actor string) error {
	if !to.Valid() {
		return fmt.Errorf("invalid status %q", to)
	}
	if err := db.withTx(func(tx *sql.Tx) error {
		var fromStr string
		err := tx.QueryRow(`SELECT status FROM tickets WHERE id=?`, id).Scan(&fromStr)
		if err == sql.ErrNoRows {
			return ErrNotFound
		}
		if err != nil {
			return err
		}
		from := ticket.Status(fromStr)
		if !ticket.CanTransition(from, to) {
			return fmt.Errorf("%w: %s -> %s", ErrIllegalTransition, from, to)
		}
		if _, err := tx.Exec(`UPDATE tickets SET status=?, updated_at=? WHERE id=?`,
			string(to), encoding.Now(), id); err != nil {
			return err
		}
		return appendAudit(tx, actor, "ticket.transitioned", "ticket", id, map[string]any{
			"from": from,
			"to":   to,
		})
	}); err != nil {
		return err
	}
	return db.writeThrough(id)
}

// TriageTicket classifies a node as leaf or composite, appending a
// ticket.triaged audit event. This is the Phase 1 store primitive; agent-driven
// decomposition is Phase 2.
func (db *DB) TriageTicket(id string, nt ticket.NodeType, actor string) error {
	if nt != ticket.NodeLeaf && nt != ticket.NodeComposite {
		return fmt.Errorf("invalid node type %q (want leaf or composite)", nt)
	}
	if err := db.withTx(func(tx *sql.Tx) error {
		res, err := tx.Exec(`UPDATE tickets SET node_type=?, updated_at=? WHERE id=?`,
			string(nt), encoding.Now(), id)
		if err != nil {
			return err
		}
		if n, _ := res.RowsAffected(); n == 0 {
			return ErrNotFound
		}
		return appendAudit(tx, actor, "ticket.triaged", "ticket", id, map[string]any{
			"node_type": nt,
		})
	}); err != nil {
		return err
	}
	return db.writeThrough(id)
}
