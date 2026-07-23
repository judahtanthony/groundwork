package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"groundwork/internal/approval"
	"groundwork/internal/store/sqlite"
)

func writeRunTranscript(t *testing.T, srv *Server, runID string, content []byte) {
	t.Helper()
	dir := filepath.Join(srv.proj.RunsDir(), runID)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "events.ndjson"), content, 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestRunTranscriptSkipsMalformedTornAndOversizedLines(t *testing.T) {
	srv, db := newTestServer(t)
	r := startRun(t, db, "transcript")

	var log bytes.Buffer
	log.WriteString(`{"time":"2026-07-22T10:00:00Z","type":"working","message":"first"}` + "\n")
	log.WriteString("not-json\n")
	log.WriteString(`{"time":"2026-07-22T10:01:00Z","type":"output","message":"`)
	log.WriteString(strings.Repeat("x", maxRunTranscriptLine+1))
	log.WriteString(`"}` + "\n")
	log.WriteString(`{"time":"2026-07-22T10:02:00Z","type":"plan_update","message":"second"}` + "\n")
	log.WriteString(`{"time":"2026-07-22T10:03:00Z","type":"output","message":`)
	writeRunTranscript(t, srv, r.ID, log.Bytes())

	var events []runTranscriptEvent
	if code := get(t, srv, "/api/v1/runs/"+r.ID+"/events", &events); code != http.StatusOK {
		t.Fatalf("events status = %d, want 200", code)
	}
	if len(events) != 2 {
		t.Fatalf("events = %+v, want two successfully parsed events", events)
	}
	if events[0].Message != "first" || events[1].Message != "second" {
		t.Errorf("messages = %q, %q", events[0].Message, events[1].Message)
	}
}

func TestRunTranscriptMissingIsEmpty(t *testing.T) {
	srv, db := newTestServer(t)
	r := startRun(t, db, "no log")
	var events []runTranscriptEvent
	if code := get(t, srv, "/api/v1/runs/"+r.ID+"/events", &events); code != http.StatusOK || len(events) != 0 {
		t.Fatalf("status=%d events=%+v, want 200 empty", code, events)
	}
}

func TestRunTranscriptGenuineReadErrorIsReported(t *testing.T) {
	srv, db := newTestServer(t)
	r := startRun(t, db, "bad file")
	path := filepath.Join(srv.proj.RunsDir(), r.ID, "events.ndjson")
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatal(err)
	}
	if code := get(t, srv, "/api/v1/runs/"+r.ID+"/events", nil); code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500 for a genuine read error", code)
	}
}

func TestRunDetailIncludesPlanFilesValidationsMetricsAndApproval(t *testing.T) {
	srv, db := newTestServer(t)
	r := startRun(t, db, "detail")
	writeRunTranscript(t, srv, r.ID, []byte(
		`{"time":"2026-07-22T10:00:00Z","type":"plan_update","message":"Implement the API"}`+"\n"))
	if err := db.SetRunChangedFiles(r.ID, []string{"web/src/main.ts", "internal/server/run_detail.go"}); err != nil {
		t.Fatal(err)
	}
	if _, err := db.RecordValidation(sqlite.ValidationResult{
		TicketID: r.TicketID, RunID: r.ID, Name: "go test", Status: sqlite.ValidationPass,
	}); err != nil {
		t.Fatal(err)
	}
	linked, err := db.CreateApproval(sqlite.CreateApprovalParams{
		RunID: r.ID, TicketID: r.TicketID, Type: approval.TypeExecute,
		RiskClass: "low", Summary: "Allow execution", Status: approval.StatusPending,
		RequestedByActor: "ai.codex.default",
	})
	if err != nil {
		t.Fatal(err)
	}

	var raw map[string]json.RawMessage
	if code := get(t, srv, "/api/v1/runs/"+r.ID, &raw); code != http.StatusOK {
		t.Fatalf("detail status = %d, want 200", code)
	}
	var detail runDetailResponse
	body, _ := json.Marshal(raw)
	if err := json.Unmarshal(body, &detail); err != nil {
		t.Fatal(err)
	}
	if detail.ID != r.ID || len(detail.Plan) != 1 || len(detail.ChangedFiles) != 2 || len(detail.Validations) != 1 {
		t.Fatalf("detail missing evidence: %+v", detail)
	}
	if detail.Approval == nil || detail.Approval.ID != linked.ID {
		t.Fatalf("approval = %+v, want %s", detail.Approval, linked.ID)
	}
	for _, field := range []string{"input_tokens", "output_tokens", "total_tokens"} {
		if _, ok := raw[field]; !ok {
			t.Errorf("detail missing metric %q", field)
		}
	}
}
