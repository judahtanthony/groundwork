package sqlite

import (
	"path/filepath"
	"testing"

	"groundwork/internal/decision"
	"groundwork/internal/ticket"
)

// TestDetectFileDivergenceFlagsUnexportedMutation proves that a durable mutation
// reaching SQLite but not files (simulating a crash between commit and sidecar
// write) is surfaced as recovery_needed rather than silently trusting SQLite
// (T-1059, ADR 0053).
func TestDetectFileDivergenceFlagsUnexportedMutation(t *testing.T) {
	dir := t.TempDir()
	ticketsDir := filepath.Join(dir, "tickets")

	db := openTestDB(t)
	db.SetExportDir(ticketsDir)

	tk := &ticket.Ticket{Title: "work", NodeType: ticket.NodeLeaf, Status: ticket.StatusTodo, WorkType: "technical_implementation"}
	if err := db.CreateTicket(tk, "tester"); err != nil { // writes sidecar
		t.Fatal(err)
	}

	// No divergence right after a write-through.
	rep, err := db.DetectFileDivergence()
	if err != nil {
		t.Fatal(err)
	}
	if len(rep.Diverged) != 0 {
		t.Fatalf("unexpected divergence: %v", rep.Diverged)
	}

	// Simulate an unexported durable mutation: change SQLite directly, bypassing
	// write-through, so the sidecar is now stale.
	if _, err := db.Exec(`UPDATE tickets SET status=? WHERE id=?`, string(ticket.StatusInProgress), tk.ID); err != nil {
		t.Fatal(err)
	}

	rep, err = db.DetectFileDivergence()
	if err != nil {
		t.Fatal(err)
	}
	if len(rep.Diverged) != 1 || rep.Diverged[0] != tk.ID {
		t.Fatalf("diverged = %v, want [%s]", rep.Diverged, tk.ID)
	}

	// recovery_needed was appended to the node's sidecar (durable, non-destructive).
	recs, ok, err := decision.Read(ticketsDir, tk.ID)
	if err != nil || !ok {
		t.Fatalf("read sidecar: ok=%v err=%v", ok, err)
	}
	if len(recs) != 1 || recs[0].EventType != decision.EventRecoveryNeeded {
		t.Fatalf("expected one recovery_needed record, got %+v", recs)
	}

	// Idempotent: a second detection does not append a duplicate flag.
	if _, err := db.DetectFileDivergence(); err != nil {
		t.Fatal(err)
	}
	recs, _, _ = decision.Read(ticketsDir, tk.ID)
	if len(recs) != 1 {
		t.Fatalf("recovery_needed duplicated: %d records", len(recs))
	}

	// The flag is in the store, so a later write-through (which rewrites the sidecar
	// from store state) PRESERVES it rather than erasing it (review finding #4).
	stored, err := db.ListDecisions(tk.ID)
	if err != nil || len(stored) != 1 || stored[0].EventType != decision.EventRecoveryNeeded {
		t.Fatalf("recovery_needed not durable in store: %+v err=%v", stored, err)
	}
	if err := db.TransitionTicket(tk.ID, ticket.StatusBlocked, "tester"); err != nil { // triggers write-through
		t.Fatal(err)
	}
	after, ok, err := decision.Read(ticketsDir, tk.ID)
	if err != nil || !ok || len(after) != 1 || after[0].EventType != decision.EventRecoveryNeeded {
		t.Fatalf("recovery_needed erased by write-through: ok=%v recs=%+v err=%v", ok, after, err)
	}
}
