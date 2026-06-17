package sqlite

import (
	"groundwork/internal/ticket"
)

// DependenciesSatisfied reports whether every node that id depends on is done.
// A node with no dependencies is trivially satisfied (ADR 0010).
func (db *DB) DependenciesSatisfied(id string) (bool, error) {
	depIDs, err := db.DependencyIDs(id)
	if err != nil {
		return false, err
	}
	for _, depID := range depIDs {
		dep, err := db.GetTicket(depID)
		if err != nil {
			return false, err
		}
		if dep.Status != ticket.StatusDone {
			return false, nil
		}
	}
	return true, nil
}

// IsEligible reports whether id is dispatchable: it is in todo and all of its
// dependencies are satisfied (docs/architecture/work-tree.md).
func (db *DB) IsEligible(id string) (bool, error) {
	t, err := db.GetTicket(id)
	if err != nil {
		return false, err
	}
	if t.Status != ticket.StatusTodo {
		return false, nil
	}
	return db.DependenciesSatisfied(id)
}

// ListEligible returns all eligible nodes, ordered by id. Eligibility is
// recomputed on each call, so it reflects dependencies completing over time.
func (db *DB) ListEligible() ([]*ticket.Ticket, error) {
	all, err := db.ListTickets()
	if err != nil {
		return nil, err
	}
	out := []*ticket.Ticket{}
	for _, t := range all {
		if t.Status != ticket.StatusTodo {
			continue
		}
		ok, err := db.DependenciesSatisfied(t.ID)
		if err != nil {
			return nil, err
		}
		if ok {
			out = append(out, t)
		}
	}
	return out, nil
}
