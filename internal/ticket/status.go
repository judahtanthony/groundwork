package ticket

// Status is a work-node lifecycle state (docs/architecture/state-model.md).
type Status string

const (
	StatusBacklog    Status = "backlog"
	StatusTodo       Status = "todo"
	StatusInProgress Status = "in_progress"
	StatusBlocked    Status = "blocked"
	StatusReview     Status = "review"
	StatusRework     Status = "rework"
	StatusApproved   Status = "approved"
	StatusLanding    Status = "landing"
	StatusDone       Status = "done"
	StatusCancelled  Status = "cancelled"
)

// AllStatuses lists every valid status in lifecycle order.
var AllStatuses = []Status{
	StatusBacklog, StatusTodo, StatusInProgress, StatusBlocked, StatusReview,
	StatusRework, StatusApproved, StatusLanding, StatusDone, StatusCancelled,
}

// Valid reports whether s is a recognized status.
func (s Status) Valid() bool {
	for _, v := range AllStatuses {
		if s == v {
			return true
		}
	}
	return false
}

// Terminal reports whether s is an end state.
func (s Status) Terminal() bool {
	return s == StatusDone || s == StatusCancelled
}

// legalTransitions is the Phase 1 manual transition map (ADR 0022). It encodes
// which transitions are structurally legal; gate-controlled transitions
// (review→approved→landing, rework cycles) are present here as valid edges, but
// their authorization (approvals, validation, reversibility) is Phase 2.
var legalTransitions = map[Status][]Status{
	StatusBacklog:    {StatusTodo, StatusCancelled},
	StatusTodo:       {StatusBacklog, StatusInProgress, StatusBlocked, StatusCancelled},
	StatusInProgress: {StatusBlocked, StatusReview, StatusDone, StatusCancelled},
	StatusBlocked:    {StatusTodo, StatusInProgress, StatusCancelled},
	StatusReview:     {StatusApproved, StatusRework, StatusCancelled},
	StatusRework:     {StatusInProgress, StatusReview, StatusCancelled},
	StatusApproved:   {StatusLanding, StatusCancelled},
	StatusLanding:    {StatusDone, StatusBlocked, StatusCancelled},
	StatusDone:       {},
	StatusCancelled:  {},
}

// CanTransition reports whether from -> to is a legal manual transition. A
// no-op transition (from == to) is allowed.
func CanTransition(from, to Status) bool {
	if from == to {
		return true
	}
	for _, allowed := range legalTransitions[from] {
		if allowed == to {
			return true
		}
	}
	return false
}
