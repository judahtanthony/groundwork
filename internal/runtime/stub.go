package runtime

import "context"

// Stub is a records-only Runtime: it emits synthetic lifecycle events and writes
// no files, so the scheduler → run → events → gate → landing loop can be tested
// without Codex (ADR 0027). Phase 4 replaces it with the real Codex adapter
// implementing the same interface.
type Stub struct{}

// Name identifies the stub runtime.
func (Stub) Name() string { return "stub" }

// Run emits a fixed lifecycle sequence and returns. It honors ctx cancellation
// between events and never touches the filesystem.
func (Stub) Run(ctx context.Context, spec Spec, sink Sink) (Result, error) {
	events := []Event{
		{Type: "claimed", Message: "claimed " + spec.TicketID},
		{Type: "working", Message: "working"},
		{Type: "produced", Message: "produced records (no code)"},
		{Type: "awaiting_gate", Message: "awaiting approval gate"},
	}
	for _, ev := range events {
		if err := ctx.Err(); err != nil {
			return Result{Status: "interrupted"}, err
		}
		if sink != nil {
			sink(ev)
		}
	}
	return Result{Status: "produced", LastMessage: "produced records (no code)"}, nil
}
