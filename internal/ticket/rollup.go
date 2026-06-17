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

	allDone := true
	anyBlocked := false
	anyActive := false
	anyUnfinished := false
	for _, c := range children {
		done := c.Status == StatusDone || c.Status == StatusCancelled
		if !done {
			allDone = false
			anyUnfinished = true
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
	case allDone:
		out.Status = StatusDone
	case anyBlocked:
		out.Status = StatusBlocked
	case anyActive:
		out.Status = StatusInProgress
	case anyUnfinished:
		out.Status = StatusTodo
	default:
		out.Status = self
	}
	return out
}
