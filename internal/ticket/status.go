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

// DependencyMet reports whether a dependency in the given status counts as
// satisfied for eligibility. Only done satisfies: a cancelled prerequisite is a
// scope-change signal requiring an explicit re-plan, not auto-satisfaction (see
// ADR 0024). This is the single satisfaction predicate shared by the eligibility
// query and the claim path.
func DependencyMet(s Status) bool {
	return s == StatusDone
}

// CanTransition reports whether from -> to is a legal manual transition. A
// no-op transition (from == to) is allowed.
//
// There is intentionally no in_progress/blocked -> rework edge: rework is for
// failed review with actionable feedback, while newly discovered scope becomes
// new dependent nodes or an upward re-plan (ADR 0022, ADR 0023). The escalation/
// re-plan cascade that drives sibling rework is human-gated and deferred to
// Phase 2.
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
