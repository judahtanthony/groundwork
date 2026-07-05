package scheduler

import (
	"bufio"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"groundwork/internal/run"
	"groundwork/internal/runtime"
)

// TestRunEventsStreamedToStoreAndJSONL proves a dispatched run's events persist
// in SQLite and append to a per-run events.ndjson, and that the run record
// carries actor_id + runtime/model metadata (T-0503, ADR 0027).
func TestRunEventsStreamedToStoreAndJSONL(t *testing.T) {
	db := newDB(t)
	tk := createTodo(t, db, "technical_implementation")

	cfg := testConfig()
	cfg.RunLogDir = filepath.Join(t.TempDir(), "runs")
	cfg.Model = "claude-test"
	s := New(db, allowCodexPolicy(), testRegistry(), runtime.Stub{}, nil, cfg)

	if _, err := s.Tick(context.Background()); err != nil {
		t.Fatalf("Tick: %v", err)
	}
	s.Wait()

	runs, _ := db.ListRunsForTicket(tk.ID)
	if len(runs) != 1 || runs[0].Status != string(run.StatusCompleted) {
		t.Fatalf("runs = %+v", runs)
	}
	r := runs[0]

	// Run record metadata (ADR 0027).
	if r.ActorID == "" || r.Runtime == "" {
		t.Fatalf("run missing actor/runtime metadata: %+v", r)
	}
	if r.Model != "claude-test" {
		t.Errorf("run model = %q, want claude-test", r.Model)
	}

	// SQLite projection has the events.
	stored, _ := db.ListRunEvents(r.ID)
	if len(stored) < 4 {
		t.Errorf("stored events = %d, want >= 4", len(stored))
	}

	// events.ndjson was appended locally with one canonical JSON line per event.
	path := filepath.Join(cfg.RunLogDir, r.ID, "events.ndjson")
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("events.ndjson missing: %v", err)
	}
	defer f.Close()
	var lines int
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		var line runEventLine
		if err := json.Unmarshal(sc.Bytes(), &line); err != nil {
			t.Fatalf("line %d not canonical JSON: %v", lines, err)
		}
		if line.Type == "" || line.Time == "" {
			t.Errorf("event line missing type/time: %+v", line)
		}
		lines++
	}
	if lines < 4 {
		t.Errorf("events.ndjson lines = %d, want >= 4", lines)
	}
}
