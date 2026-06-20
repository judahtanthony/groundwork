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
}

// Result summarizes a completed attempt.
type Result struct {
	Status      string // outcome hint, e.g. "produced"
	LastMessage string
}

// Runtime executes one node attempt, emitting events to sink until done or ctx
// is cancelled.
type Runtime interface {
	Name() string
	Run(ctx context.Context, spec Spec, sink Sink) (Result, error)
}
