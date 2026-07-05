package scheduler

import (
	"context"
	"testing"

	"groundwork/internal/runtime"
	"groundwork/internal/ticket"
)

// blockedRuntime returns a blocked outcome with a handoff summary.
type blockedRuntime struct{}

func (blockedRuntime) Name() string { return "blocked-test" }
func (blockedRuntime) Run(ctx context.Context, spec runtime.Spec, sink runtime.Sink) (runtime.Result, error) {
	if sink != nil {
		sink(runtime.Event{Type: "working"})
	}
	return runtime.Result{
		Status:         runtime.OutcomeBlocked,
		Statement:      "which serialization format?",
		HandoffSummary: "blocked on the serialization decision",
	}, nil
}

// TestBlockedOutcomeMovesToBlockedWithHandoff proves a blocked run moves the
// ticket to blocked (not review) and writes a durable handoff record, so capacity
// is released and a later run can resume (T-1055, ADR 0051).
func TestBlockedOutcomeMovesToBlockedWithHandoff(t *testing.T) {
	db := newDB(t)
	tk := createTodo(t, db, "technical_implementation")
	s := New(db, allowCodexPolicy(), testRegistry(), blockedRuntime{}, nil, testConfig())

	if _, err := s.Tick(context.Background()); err != nil {
		t.Fatalf("Tick: %v", err)
	}
	s.Wait()

	got, _ := db.GetTicket(tk.ID)
	if got.Status != ticket.StatusBlocked {
		t.Fatalf("status = %s, want blocked", got.Status)
	}
	// A durable blocker explains the block and carries the handoff summary.
	recs, err := db.ListDecisions(tk.ID)
	if err != nil || len(recs) != 1 {
		t.Fatalf("decisions = %+v err=%v", recs, err)
	}
	if recs[0].HandoffSummary != "blocked on the serialization decision" {
		t.Errorf("handoff = %q", recs[0].HandoffSummary)
	}
	// The lease was released (capacity returned).
	if lease, _ := db.GetLease(tk.ID); lease != nil {
		t.Error("lease not released after blocked handoff")
	}
}
