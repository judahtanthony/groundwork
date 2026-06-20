// Package run holds the run domain model: a run is one supervised attempt at a
// node (runtime-model.md). It carries a lifecycle status and a mode (planning or
// implementation) as state machines, independent of the node's ticket status
// (ADR 0027). The actual agent execution is the Phase 4 runtime; this package is
// the records/state half.
package run

// Status is a run lifecycle state (ADR 0027).
type Status string

const (
	StatusPending     Status = "pending"     // created, not yet executing
	StatusRunning     Status = "running"     // actively executing
	StatusPaused      Status = "paused"      // paused at a safe boundary
	StatusInterrupted Status = "interrupted" // worker lost; set only by recovery
	StatusCompleted   Status = "completed"   // finished normally
	StatusCancelled   Status = "cancelled"   // intentionally stopped
)

// AllStatuses lists every run status.
var AllStatuses = []Status{
	StatusPending, StatusRunning, StatusPaused, StatusInterrupted, StatusCompleted, StatusCancelled,
}

// Valid reports whether s is a recognized run status.
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
	return s == StatusCompleted || s == StatusCancelled
}

// Mode is the run mode (runtime-model.md).
type Mode string

const (
	ModePlanning       Mode = "planning"       // triage/decompose, produces a proposal
	ModeImplementation Mode = "implementation" // executes a leaf in a worktree
)

// Valid reports whether m is a recognized mode.
func (m Mode) Valid() bool {
	return m == ModePlanning || m == ModeImplementation
}

// legalTransitions encodes the run lifecycle. interrupted is reachable from any
// live state (set by recovery, not by clients) and can resume to running.
var legalTransitions = map[Status][]Status{
	StatusPending:     {StatusRunning, StatusCancelled, StatusInterrupted},
	StatusRunning:     {StatusPaused, StatusCompleted, StatusCancelled, StatusInterrupted},
	StatusPaused:      {StatusRunning, StatusCancelled, StatusInterrupted},
	StatusInterrupted: {StatusRunning, StatusCancelled},
	StatusCompleted:   {},
	StatusCancelled:   {},
}

// CanTransition reports whether from -> to is a legal run transition. A no-op
// (from == to) is allowed.
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
