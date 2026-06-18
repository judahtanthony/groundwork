package sqlite

import (
	"fmt"

	"groundwork/internal/ticket"
)

// ListChildren returns the direct children of parentID, ordered by id.
func (db *DB) ListChildren(parentID string) ([]*ticket.Ticket, error) {
	rows, err := db.Query(`SELECT `+ticketColumns+` FROM tickets WHERE parent_id=? ORDER BY id`, parentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanTickets(rows)
}

// Ancestors returns id's ancestor spine ordered root-first (excluding id
// itself). It guards against malformed parent cycles with a depth limit.
func (db *DB) Ancestors(id string) ([]*ticket.Ticket, error) {
	cur, err := db.GetTicket(id)
	if err != nil {
		return nil, err
	}
	var chain []*ticket.Ticket
	seen := map[string]bool{cur.ID: true}
	for cur.ParentID != "" {
		if seen[cur.ParentID] {
			return nil, fmt.Errorf("parent cycle detected at %s", cur.ParentID)
		}
		seen[cur.ParentID] = true
		parent, err := db.GetTicket(cur.ParentID)
		if err != nil {
			return nil, err
		}
		chain = append(chain, parent)
		cur = parent
	}
	// Reverse to root-first.
	for i, j := 0, len(chain)-1; i < j; i, j = i+1, j-1 {
		chain[i], chain[j] = chain[j], chain[i]
	}
	return chain, nil
}
