package scheduler

import (
	"sort"

	"groundwork/internal/ticket"
)

// orderByValue sorts the eligible set by value (ADR 0039), replacing FIFO-by-id.
// It is the single `score` seam: today the score is the priority path computed
// below; the future multi-signal value model (value/effort/risk/confidence/depth)
// replaces this function without touching the scheduler loop.
func (s *Scheduler) orderByValue(nodes []*ticket.Ticket) {
	key := make(map[string][]levelKey, len(nodes))
	for _, n := range nodes {
		key[n.ID] = s.priorityPath(n)
	}
	sort.SliceStable(nodes, func(i, j int) bool {
		return comparePath(key[nodes[i].ID], key[nodes[j].ID]) < 0
	})
}

// levelKey is one node's contribution to its path's sort key: its effective
// priority and its id.
type levelKey struct {
	prio float64
	id   string
}

// priorityPath builds the root→node path as (effective priority, id) per level.
// Priority is sibling-scoped: because siblings share their ancestor prefix, a
// node's own priority only discriminates it from its siblings, and cross-subtree
// order is resolved at the ancestor divergence point (ADR 0039). On a lookup
// error the path degrades to the node alone, so ordering falls back to id.
func (s *Scheduler) priorityPath(n *ticket.Ticket) []levelKey {
	ancestors, err := s.db.Ancestors(n.ID)
	if err != nil {
		ancestors = nil
	}
	path := make([]levelKey, 0, len(ancestors)+1)
	for _, a := range ancestors {
		path = append(path, levelKey{a.EffectivePriority(), a.ID})
	}
	return append(path, levelKey{n.EffectivePriority(), n.ID})
}

// comparePath orders two paths lexicographically: at each level, higher priority
// runs first, then lower id (DFS/FIFO). A shorter path (an ancestor) sorts first
// when it is a prefix of the other. Returns <0 if a should run before b.
func comparePath(a, b []levelKey) int {
	for i := 0; i < len(a) && i < len(b); i++ {
		if a[i].prio != b[i].prio {
			if a[i].prio > b[i].prio {
				return -1
			}
			return 1
		}
		if a[i].id != b[i].id {
			if a[i].id < b[i].id {
				return -1
			}
			return 1
		}
	}
	return len(a) - len(b)
}
