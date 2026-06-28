package scheduler

import (
	"context"
	"testing"

	"groundwork/internal/actor"
	"groundwork/internal/runtime"
)

// fakeGate returns a fixed claim decision and records the calls.
type fakeGate struct {
	decision ClaimDecision
	calls    int
}

func (g *fakeGate) AuthorizeAIClaim(nodeID, action, workType string, a *actor.Actor) (ClaimDecision, error) {
	g.calls++
	return g.decision, nil
}

func TestEnvelopeGateAllowsDispatch(t *testing.T) {
	db := newDB(t)
	tk := createTodo(t, db, "technical_implementation")
	// Trust policy denies (empty); only the gate authorizes — proving the scheduler
	// routes through the gate, not the trust-only path.
	s := New(db, allowCodexPolicy(), testRegistry(), runtime.Stub{}, nil, testConfig())
	gate := &fakeGate{decision: ClaimAllow}
	s.SetEnvelopeGate(gate)

	started, err := s.Tick(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	s.Wait()
	if started != 1 {
		t.Fatalf("started = %d, want 1", started)
	}
	if gate.calls == 0 {
		t.Error("gate was not consulted")
	}
	if runs, _ := db.ListRunsForTicket(tk.ID); len(runs) != 1 {
		t.Fatalf("runs = %d, want 1", len(runs))
	}
}

func TestEnvelopeGateDenyBlocksDispatch(t *testing.T) {
	db := newDB(t)
	createTodo(t, db, "technical_implementation")
	s := New(db, allowCodexPolicy(), testRegistry(), runtime.Stub{}, nil, testConfig())
	s.SetEnvelopeGate(&fakeGate{decision: ClaimDeny})

	started, err := s.Tick(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if started != 0 {
		t.Fatalf("started = %d, want 0 (no envelope authorizes the claim)", started)
	}
}

func TestEnvelopeGateExceptionDoesNotDispatch(t *testing.T) {
	db := newDB(t)
	tk := createTodo(t, db, "technical_implementation")
	s := New(db, allowCodexPolicy(), testRegistry(), runtime.Stub{}, nil, testConfig())
	s.SetEnvelopeGate(&fakeGate{decision: ClaimException})

	started, err := s.Tick(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if started != 0 {
		t.Fatalf("started = %d, want 0 (boundary crossing raises an exception, not a run)", started)
	}
	// The node stays unclaimed (no run) pending the human exception decision.
	if runs, _ := db.ListRunsForTicket(tk.ID); len(runs) != 0 {
		t.Fatalf("runs = %d, want 0", len(runs))
	}
}
