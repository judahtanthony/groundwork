package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"groundwork/internal/decision"
	"groundwork/internal/store/sqlite"
	"groundwork/internal/ticket"
)

// TestImportDecisionsSidecarRoundTrip proves a ticket's decisions.ndjson sidecar
// is projected into the store on import and re-exports byte-for-byte (T-1053,
// ADR 0051/0053): the cold-rebuild invariant for durable decision records.
func TestImportDecisionsSidecarRoundTrip(t *testing.T) {
	dir := t.TempDir()
	tk := &ticket.Ticket{ID: "T-1234", Kind: "ticket", Title: "Blocked work", Status: ticket.StatusBlocked}
	writeExport(t, dir, tk, nil)

	recs := []decision.Record{
		{ID: "D-0001", Sequence: 1, EventType: decision.EventApprovalRequested, TicketID: "T-1234",
			Status: decision.StatusPending, RequestedBy: "ai.codex.default", RequestedAt: "2026-06-24T15:00:00Z",
			Statement: "Accept the proposed children?"},
		{ID: "D-0002", Sequence: 2, EventType: decision.EventRecoveryNeeded, TicketID: "T-1234",
			Status: decision.StatusPending, HandoffSummary: "no durable blocker after rebuild"},
	}
	if err := decision.Write(dir, "T-1234", recs); err != nil {
		t.Fatal(err)
	}
	wantBytes, err := os.ReadFile(decision.Path(dir, "T-1234"))
	if err != nil {
		t.Fatal(err)
	}

	db, err := sqlite.Open(filepath.Join(t.TempDir(), "state.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err := db.Migrate(); err != nil {
		t.Fatal(err)
	}
	if _, err := importExports(db, dir); err != nil {
		t.Fatalf("import: %v", err)
	}

	got, err := db.ListDecisions("T-1234")
	if err != nil || len(got) != 2 {
		t.Fatalf("projected decisions n=%d err=%v", len(got), err)
	}
	pending, err := db.ListPendingDecisions()
	if err != nil || len(pending) != 2 {
		t.Fatalf("pending n=%d err=%v", len(pending), err)
	}

	// Re-export the sidecar from the rebuilt store; bytes must match the committed file.
	rebuilt, err := db.ListDecisions("T-1234")
	if err != nil {
		t.Fatal(err)
	}
	gotBytes, err := decision.Encode(rebuilt)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(gotBytes, wantBytes) {
		t.Fatalf("decisions sidecar not byte-stable after rebuild:\n--- got ---\n%s\n--- want ---\n%s", gotBytes, wantBytes)
	}
}
