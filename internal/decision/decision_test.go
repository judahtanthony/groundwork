package decision

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func boolPtr(b bool) *bool { return &b }

// sampleRecords covers one of each contract event class: pending input,
// approval, decision, rework, and recovery.
func sampleRecords() []Record {
	return []Record{
		{
			ID: "D-0001", Sequence: 1, EventType: EventInputRequested, TicketID: "T-1234",
			RunID: "R-1", Status: StatusPending, RequestedBy: "ai.codex.default",
			RequestedAt: "2026-06-24T15:04:05Z", Statement: "Which config key holds the timeout?",
		},
		{
			ID: "D-0002", Sequence: 2, EventType: EventApprovalRequested, TicketID: "T-1234",
			RunID: "R-1", RequestType: "decompose", WorkType: "technical_design",
			Status: StatusPending, RequestedBy: "ai.codex.default", RequestedAt: "2026-06-24T15:05:00Z",
			RequestedActor: "human.owner", RequiredRoles: []string{"owner"},
			PolicyInputs: &PolicyInputs{Action: "decompose", RiskClass: "medium", Reversible: boolPtr(true)},
			Statement:    "Accept the proposed child tickets?",
			Options:      []Option{{ID: "accept", Label: "Accept"}, {ID: "revise", Label: "Rework"}},
			Recommendation: "accept", RelatedFiles: []string{".groundwork/tickets/T-1234/ticket.md"},
		},
		{
			ID: "D-0003", Sequence: 3, EventType: EventDecisionRequested, TicketID: "T-1234",
			Status: StatusPending, RequestedBy: "ai.codex.default", RequestedAt: "2026-06-24T15:06:00Z",
			Statement: "Should we adopt the new schema?", DependsOn: []string{"T-2000"},
			HandoffSummary: "Consequential schema choice; routed to a decision ticket.",
		},
		{
			ID: "D-0004", Sequence: 4, EventType: EventReworkRequested, TicketID: "T-1234",
			Status: StatusRejected, DecidedBy: "human.owner", DecidedAt: "2026-06-24T16:00:00Z",
			Statement: "Tests missing for the failure path.", FollowUp: []string{"add failure-path test"},
		},
		{
			ID: "D-0005", Sequence: 5, EventType: EventRecoveryNeeded, TicketID: "T-1234",
			Status: StatusPending, RequestedAt: "2026-06-24T17:00:00Z",
			HandoffSummary: "Blocked ticket has no durable blocker after rebuild.",
		},
	}
}

func TestEncodeDecodeRoundTrip(t *testing.T) {
	recs := sampleRecords()
	data, err := Encode(recs)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}
	got, err := Decode(data)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(got) != len(recs) {
		t.Fatalf("record count: got %d want %d", len(got), len(recs))
	}
	// Re-encoding the decoded records must reproduce identical bytes (canonical
	// byte-stability — the cold-rebuild invariant).
	again, err := Encode(got)
	if err != nil {
		t.Fatalf("re-encode: %v", err)
	}
	if !bytes.Equal(data, again) {
		t.Fatalf("re-encode not byte-stable:\nfirst:\n%s\nsecond:\n%s", data, again)
	}
}

func TestEncodeDeterministicOrder(t *testing.T) {
	recs := sampleRecords()
	// Shuffle the input order; Encode must sort by sequence and produce the same
	// bytes regardless of caller order.
	shuffled := []Record{recs[3], recs[0], recs[4], recs[1], recs[2]}
	a, err := Encode(recs)
	if err != nil {
		t.Fatal(err)
	}
	b, err := Encode(shuffled)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(a, b) {
		t.Fatalf("encode order-dependent:\nordered:\n%s\nshuffled:\n%s", a, b)
	}
}

func TestEncodeOneLinePerRecord(t *testing.T) {
	data, err := Encode(sampleRecords())
	if err != nil {
		t.Fatal(err)
	}
	if n := bytes.Count(data, []byte("\n")); n != 5 {
		t.Fatalf("expected 5 lines, got %d", n)
	}
	if !bytes.HasSuffix(data, []byte("\n")) {
		t.Fatal("expected trailing newline")
	}
}

func TestValidateRejectsMissingFields(t *testing.T) {
	cases := map[string]Record{
		"no event type": {TicketID: "T-1", Status: StatusPending, Statement: "x"},
		"no ticket":     {EventType: EventInputRequested, Status: StatusPending, Statement: "x"},
		"no status":     {EventType: EventInputRequested, TicketID: "T-1", Statement: "x"},
		"no text":       {EventType: EventInputRequested, TicketID: "T-1", Status: StatusPending},
	}
	for name, r := range cases {
		if err := r.Validate(); err == nil {
			t.Errorf("%s: expected validation error", name)
		}
	}
}

func TestDecodeRejectsUnknownFields(t *testing.T) {
	if _, err := Decode([]byte(`{"event_type":"input_requested","ticket_id":"T-1","status":"pending","statement":"x","bogus":1}` + "\n")); err == nil {
		t.Fatal("expected error on unknown field")
	}
}

func TestDecodeToleratesBlankLines(t *testing.T) {
	line := `{"sequence":1,"event_type":"input_requested","ticket_id":"T-1","status":"pending","statement":"x"}`
	recs, err := Decode([]byte("\n" + line + "\n\n"))
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(recs) != 1 {
		t.Fatalf("got %d records", len(recs))
	}
}

func TestWriteReadSidecar(t *testing.T) {
	dir := t.TempDir()
	if err := Write(dir, "T-1234", sampleRecords()); err != nil {
		t.Fatalf("write: %v", err)
	}
	recs, ok, err := Read(dir, "T-1234")
	if err != nil || !ok {
		t.Fatalf("read: ok=%v err=%v", ok, err)
	}
	if len(recs) != 5 {
		t.Fatalf("got %d records", len(recs))
	}
}

func TestWriteEmptyRemovesSidecar(t *testing.T) {
	dir := t.TempDir()
	if err := Write(dir, "T-1", sampleRecords()); err != nil {
		t.Fatal(err)
	}
	path := Path(dir, "T-1")
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("sidecar should exist: %v", err)
	}
	// Writing no records removes the optional sidecar instead of leaving an empty file.
	if err := Write(dir, "T-1", nil); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("sidecar should be removed, stat err=%v", err)
	}
	// Read of an absent sidecar reports ok=false, no error.
	_, ok, err := Read(dir, "T-1")
	if err != nil || ok {
		t.Fatalf("read absent: ok=%v err=%v", ok, err)
	}
}

func TestReadMissingTicketDir(t *testing.T) {
	dir := t.TempDir()
	_, ok, err := Read(filepath.Join(dir, "nope"), "T-9")
	if err != nil || ok {
		t.Fatalf("expected ok=false err=nil, got ok=%v err=%v", ok, err)
	}
}
