package server

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"

	"groundwork/internal/store/sqlite"
)

// maxRunTranscriptLine bounds one event while leaving enough room for normal
// model/tool messages. The split function discards a larger line and resumes at
// the next newline instead of making transcript content an API failure.
const maxRunTranscriptLine = 1024 * 1024

// runTranscriptEvent is the events.ndjson representation exposed to API
// clients. The field names retain compatibility with the original SQLite event
// response while adding the per-event message that only the durable log holds.
type runTranscriptEvent struct {
	ID        int64  `json:"id"`
	RunID     string `json:"run_id"`
	EventType string `json:"event_type"`
	Message   string `json:"message,omitempty"`
	Payload   string `json:"payload"`
	CreatedAt string `json:"created_at"`
}

type runEventLine struct {
	Time    string         `json:"time"`
	Type    string         `json:"type"`
	Message string         `json:"message,omitempty"`
	Payload map[string]any `json:"payload,omitempty"`
}

type runDetailResponse struct {
	*sqlite.Run
	Plan         []runTranscriptEvent       `json:"plan"`
	ChangedFiles []string                   `json:"changed_files"`
	Validations  []*sqlite.ValidationResult `json:"validations"`
	Approval     *sqlite.Approval           `json:"approval,omitempty"`
	Cost         *float64                   `json:"cost,omitempty"`
}

// readRunTranscript reads ignored per-run evidence from the project run-log
// directory. Missing logs are a normal empty state. Bad JSON, a partial final
// append, and individual over-cap lines are content problems and are skipped;
// only an actual filesystem/read failure is returned.
func (s *Server) readRunTranscript(runID string) ([]runTranscriptEvent, error) {
	path := filepath.Join(s.proj.RunsDir(), runID, "events.ndjson")
	f, err := os.Open(path)
	if errors.Is(err, os.ErrNotExist) {
		return []runTranscriptEvent{}, nil
	}
	if err != nil {
		return nil, err
	}
	defer f.Close()

	events := []runTranscriptEvent{}
	scanner := bufio.NewScanner(f)
	// scanRunTranscriptLines discards over-cap input before Scanner reaches this
	// maximum; the extra byte lets it distinguish an exactly-at-cap line.
	scanner.Buffer(make([]byte, 64*1024), maxRunTranscriptLine+1)
	scanner.Split(scanRunTranscriptLines(maxRunTranscriptLine))
	for scanner.Scan() {
		var line runEventLine
		if err := json.Unmarshal(scanner.Bytes(), &line); err != nil || line.Type == "" || line.Time == "" {
			continue
		}
		payload := "{}"
		if line.Payload != nil {
			if encoded, err := json.Marshal(line.Payload); err == nil {
				payload = string(encoded)
			}
		}
		events = append(events, runTranscriptEvent{
			ID: int64(len(events) + 1), RunID: runID, EventType: line.Type,
			Message: line.Message, Payload: payload, CreatedAt: line.Time,
		})
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return events, nil
}

// scanRunTranscriptLines is ScanLines with recovery for a single oversized
// record. Once the cap is crossed it consumes chunks until the next newline and
// emits no token, allowing later valid events to remain visible.
func scanRunTranscriptLines(max int) bufio.SplitFunc {
	discarding := false
	return func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		if i := bytes.IndexByte(data, '\n'); i >= 0 {
			line := data[:i]
			if discarding || len(line) > max {
				discarding = false
				return i + 1, nil, nil
			}
			return i + 1, bytes.TrimSuffix(line, []byte{'\r'}), nil
		}
		if discarding {
			if atEOF {
				discarding = false
				return len(data), nil, nil
			}
			if len(data) > max {
				return len(data), nil, nil
			}
			return 0, nil, nil
		}
		if len(data) > max {
			discarding = true
			return len(data), nil, nil
		}
		if atEOF {
			if len(data) == 0 {
				return 0, nil, nil
			}
			return len(data), bytes.TrimSuffix(data, []byte{'\r'}), nil
		}
		return 0, nil, nil
	}
}

func planEvents(events []runTranscriptEvent) []runTranscriptEvent {
	plan := []runTranscriptEvent{}
	for _, event := range events {
		kind := strings.ToLower(event.EventType)
		if strings.Contains(kind, "plan") || strings.Contains(kind, "triage") || strings.Contains(kind, "decompos") {
			plan = append(plan, event)
		}
	}
	return plan
}

func (s *Server) runValidations(runID, ticketID string) ([]*sqlite.ValidationResult, error) {
	all, err := s.db.ListValidationsForTicket(ticketID)
	if err != nil {
		return nil, err
	}
	result := []*sqlite.ValidationResult{}
	for _, validation := range all {
		// Older/manual validation records were ticket-scoped. They remain relevant
		// evidence when no run id was captured.
		if validation.RunID == "" || validation.RunID == runID {
			result = append(result, validation)
		}
	}
	return result, nil
}

func (s *Server) approvalForRun(runID string) (*sqlite.Approval, error) {
	approvals, err := s.db.ListApprovals("")
	if err != nil {
		return nil, err
	}
	for _, candidate := range approvals {
		if candidate.RunID == runID {
			return candidate, nil
		}
	}
	return nil, nil
}
