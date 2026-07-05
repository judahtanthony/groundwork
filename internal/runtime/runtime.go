// Package runtime is the agent-runtime seam (agent-runtime.md). A Runtime
// executes one supervised node attempt, emitting events as it goes. The Codex
// adapter is Phase 4; M2 ships a records-only Stub so the coordinator loop can be
// exercised end-to-end without launching an external agent (ADR 0027).
package runtime

import "context"

// Event is a runtime lifecycle/telemetry event emitted during a run.
type Event struct {
	Type    string
	Message string
	Payload map[string]any
}

// Sink receives runtime events as they occur.
type Sink func(Event)

// Spec describes the node attempt to run.
type Spec struct {
	RunID     string
	TicketID  string
	Mode      string // run.Mode value
	ActorID   string
	Runtime   string
	Model     string
	Workspace string
	// Prompt is the task instruction handed to the agent — assembled by the
	// coordinator from the ticket's durable context (acceptance, contract, blockers).
	Prompt string
}

// Run outcomes (ADR 0051). A run ends in exactly one. Produced/Completed reach
// review; Blocked/InputRequired/Escalated/Rework end the run durably so capacity
// moves elsewhere; Cancelled/Interrupted are off-ramps.
const (
	OutcomeProduced      = "produced"       // work produced, awaiting its gate (-> review)
	OutcomeCompleted     = "completed"      // nothing more to do
	OutcomeBlocked       = "blocked"        // durable blocker; needs a decision/dependency
	OutcomeInputRequired = "input_required" // a bounded clarification is needed to continue
	OutcomeEscalated     = "escalated"      // exceeded scope/affordance; needs re-plan/approval
	OutcomeRework        = "rework"         // self-identified as needing rework
	OutcomeCancelled     = "cancelled"      // cancelled by a client
	OutcomeInterrupted   = "interrupted"    // killed mid-run (crash/ctx cancel)
)

// IsBlockedOutcome reports whether an outcome should move the ticket to blocked
// with a durable handoff rather than to review (ADR 0051).
func IsBlockedOutcome(status string) bool {
	switch status {
	case OutcomeBlocked, OutcomeInputRequired, OutcomeEscalated:
		return true
	}
	return false
}

// Result summarizes a completed attempt.
type Result struct {
	Status      string // one of the Outcome* values
	LastMessage string
	// HandoffSummary explains a blocked/escalated exit so a later run can resume
	// without the original session (ADR 0051/0047). Required for blocked outcomes.
	HandoffSummary string
	// Statement is the question/decision a blocked or input-required run is waiting
	// on, recorded on the durable blocker.
	Statement string
	// ChangedFiles is the run's changed-file set (worktree diff vs base, ADR 0059).
	// It is the authoritative diff for gate inputs: validation template selection,
	// envelope file-scope, and escalation triggers. Empty for a no-change run.
	ChangedFiles []string
	// Diff is the run's unified diff against base — run evidence for human review.
	Diff string
}

// Runtime executes one node attempt, emitting events to sink until done or ctx
// is cancelled.
type Runtime interface {
	Name() string
	Run(ctx context.Context, spec Spec, sink Sink) (Result, error)
}
