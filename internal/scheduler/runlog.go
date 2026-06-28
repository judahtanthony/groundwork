package scheduler

import (
	"encoding/json"
	"os"
	"path/filepath"

	"groundwork/internal/encoding"
	"groundwork/internal/runtime"
)

// runEventLine is one canonical JSON line in a run's events.ndjson (ADR 0027).
// The JSONL log is tier-1 ignored runtime evidence under .groundwork/runs; SQLite
// holds the queryable projection (run_events).
type runEventLine struct {
	Time    string         `json:"time"`
	Type    string         `json:"type"`
	Message string         `json:"message,omitempty"`
	Payload map[string]any `json:"payload,omitempty"`
}

// appendRunEventLog appends one event to <dir>/<runID>/events.ndjson, creating
// the run directory as needed. A blank dir disables the local log (e.g. tests).
func appendRunEventLog(dir, runID string, ev runtime.Event) error {
	if dir == "" {
		return nil
	}
	runDir := filepath.Join(dir, runID)
	if err := os.MkdirAll(runDir, 0o755); err != nil {
		return err
	}
	f, err := os.OpenFile(filepath.Join(runDir, "events.ndjson"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	line, err := json.Marshal(runEventLine{
		Time: encoding.Now(), Type: ev.Type, Message: ev.Message, Payload: ev.Payload,
	})
	if err != nil {
		return err
	}
	if _, err := f.Write(append(line, '\n')); err != nil {
		return err
	}
	return nil
}
