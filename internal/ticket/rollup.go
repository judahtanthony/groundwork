package ticket

// Rollup is a parent node's derived state, computed from its descendants
// (docs/architecture/work-tree.md). It is never persisted: executable state
// belongs to leaves and parent state is always computed (ADR 0022).
type Rollup struct {
	Status     Status `json:"status"`
	HasBlocked bool   `json:"has_blocked"`
	HasActive  bool   `json:"has_active"`
}

// isActive reports whether a status represents work in motion.
func isActive(s Status) bool {
	switch s {
	case StatusInProgress, StatusReview, StatusRework, StatusApproved, StatusLanding:
		return true
	}
	return false
}

// ComputeRollup derives a node's rollup from its own status and the rollups of
// its direct children. With no children (a leaf) the node's own status is used.
// Status precedence follows work-tree.md: done > blocked > in_progress > todo.
func ComputeRollup(self Status, children []Rollup) Rollup {
	if len(children) == 0 {
		return Rollup{
			Status:     self,
			HasBlocked: self == StatusBlocked,
			HasActive:  isActive(self),
		}
	}

	allTerminal := true
	anyDone := false
	anyBlocked := false
	anyActive := false
	for _, c := range children {
		// done and cancelled are both terminal/settled (ADR 0024).
		if c.Status != StatusDone && c.Status != StatusCancelled {
			allTerminal = false
		}
		if c.Status == StatusDone {
			anyDone = true
		}
		if c.HasBlocked || c.Status == StatusBlocked {
			anyBlocked = true
		}
		if c.HasActive || isActive(c.Status) {
			anyActive = true
		}
	}

	out := Rollup{HasBlocked: anyBlocked, HasActive: anyActive}
	switch {
	case allTerminal && anyDone:
		out.Status = StatusDone
	case allTerminal:
		// All children settled but none done: everything was cancelled.
		out.Status = StatusCancelled
	case anyBlocked:
		out.Status = StatusBlocked
	case anyActive:
		out.Status = StatusInProgress
	default:
		out.Status = StatusTodo
	}
	return out
}
