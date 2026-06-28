package scheduler

import (
	"context"
	"path/filepath"
	"testing"

	"groundwork/internal/completion"
	"groundwork/internal/runtime"
)

// producingRuntime returns a produced result carrying a changed-file set.
type producingRuntime struct{}

func (producingRuntime) Name() string { return "producing-test" }
func (producingRuntime) Run(ctx context.Context, spec runtime.Spec, sink runtime.Sink) (runtime.Result, error) {
	return runtime.Result{Status: runtime.OutcomeProduced, LastMessage: "did the thing",
		ChangedFiles: []string{"main.go"}}, nil
}

// TestProducedRunWritesCompletionSummary proves a runtime-produced result gets a
// completion summary before review (T-1058, ADR 0047).
func TestProducedRunWritesCompletionSummary(t *testing.T) {
	db := newDB(t)
	tk := createTodo(t, db, "technical_implementation")
	ticketsDir := filepath.Join(t.TempDir(), "tickets")

	cfg := testConfig()
	cfg.TicketsDir = ticketsDir
	s := New(db, allowCodexPolicy(), testRegistry(), producingRuntime{}, nil, cfg)

	if _, err := s.Tick(context.Background()); err != nil {
		t.Fatalf("Tick: %v", err)
	}
	s.Wait()

	sum, ok, err := completion.Read(ticketsDir, tk.ID)
	if err != nil || !ok {
		t.Fatalf("completion summary not written: ok=%v err=%v", ok, err)
	}
	if len(sum.Changed) != 1 || sum.Changed[0] != "main.go" {
		t.Errorf("summary changed = %v, want [main.go]", sum.Changed)
	}
	if sum.Outcome != runtime.OutcomeProduced {
		t.Errorf("summary outcome = %q", sum.Outcome)
	}
	// Mirrored into SQLite for the review bundle.
	got, err := db.GetCompletionSummary(tk.ID)
	if err != nil || got == nil {
		t.Fatalf("summary not mirrored: got=%v err=%v", got, err)
	}
}

// TestProducedRunKeepsExistingSummary proves a pre-existing (human) summary is not
// clobbered by the auto-summary.
func TestProducedRunKeepsExistingSummary(t *testing.T) {
	db := newDB(t)
	tk := createTodo(t, db, "technical_implementation")
	ticketsDir := filepath.Join(t.TempDir(), "tickets")
	if err := completion.Write(ticketsDir, &completion.Summary{NodeID: tk.ID, Outcome: "human-authored"}); err != nil {
		t.Fatal(err)
	}

	cfg := testConfig()
	cfg.TicketsDir = ticketsDir
	s := New(db, allowCodexPolicy(), testRegistry(), producingRuntime{}, nil, cfg)
	if _, err := s.Tick(context.Background()); err != nil {
		t.Fatal(err)
	}
	s.Wait()

	sum, _, _ := completion.Read(ticketsDir, tk.ID)
	if sum.Outcome != "human-authored" {
		t.Errorf("auto-summary clobbered the existing one: %q", sum.Outcome)
	}
}
