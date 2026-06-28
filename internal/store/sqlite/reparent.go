package sqlite

import (
	"database/sql"
	"errors"
	"fmt"

	"groundwork/internal/encoding"
)

var (
	// ErrSelfParent rejects parenting a node under itself.
	ErrSelfParent = errors.New("a node cannot be its own parent")
	// ErrParentCycle rejects parenting a node under one of its own descendants,
	// which would create a cycle in the parent tree.
	ErrParentCycle = errors.New("reparenting under a descendant would create a cycle")
)

// Reparent moves id under newParentID, appending a ticket.reparented audit event.
// It rejects a missing target, self-parenting, and parenting under one's own
// descendant. Rollups are derived on read (status rollups recompute from the
// tree), so both the old and new parents reflect the move on the next read with
// no stored recompute (ADR 0041, T-1021). UpdateTicket deliberately leaves
// parent_id alone (ADR 0022), so this is the one path that changes parentage.
func (db *DB) Reparent(id, newParentID, actor string) error {
	if newParentID == "" {
		return fmt.Errorf("new parent id is required")
	}
	if newParentID == id {
		return ErrSelfParent
	}
	node, err := db.GetTicket(id)
	if err != nil {
		return err
	}
	if _, err := db.GetTicket(newParentID); err != nil {
		if errors.Is(err, ErrNotFound) {
			return fmt.Errorf("%w: parent %s", ErrNotFound, newParentID)
		}
		return err
	}
	// id must not be an ancestor of the new parent, or the move forms a cycle.
	ancestors, err := db.Ancestors(newParentID)
	if err != nil {
		return err
	}
	for _, a := range ancestors {
		if a.ID == id {
			return ErrParentCycle
		}
	}

	oldParent := node.ParentID
	if err := db.withTx(func(tx *sql.Tx) error {
		res, err := tx.Exec(`UPDATE tickets SET parent_id=?, updated_at=? WHERE id=?`,
			nullStr(newParentID), encoding.Now(), id)
		if err != nil {
			return err
		}
		if n, _ := res.RowsAffected(); n == 0 {
			return ErrNotFound
		}
		return appendAudit(tx, actor, "ticket.reparented", "ticket", id, map[string]any{
			"from": oldParent,
			"to":   newParentID,
		})
	}); err != nil {
		return err
	}
	return db.writeThrough(id)
}
