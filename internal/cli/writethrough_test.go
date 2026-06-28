package cli

import (
	"os"
	"path/filepath"
	"testing"

	"groundwork/internal/decision"
	"groundwork/internal/store/sqlite"
	"groundwork/internal/ticket"
)

// TestWriteThroughSurvivesPurgeRebuild proves ADR 0053: durable ticket/dependency/
// decision mutations are written through to sidecars, so deleting state.sqlite and
// rebuilding from files preserves ticket records, statuses, dependencies, and
// decision records (T-1059).
func TestWriteThroughSurvivesPurgeRebuild(t *testing.T) {
	dir := t.TempDir()
	ticketsDir := filepath.Join(dir, "tickets")

	db1, err := sqlite.Open(filepath.Join(dir, "state.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	if err := db1.Migrate(); err != nil {
		t.Fatal(err)
	}
	db1.SetExportDir(ticketsDir)

	// Representative durable mutations, each of which must write through to files.
	parent := &ticket.Ticket{Title: "epic", Kind: "epic", NodeType: ticket.NodeComposite, Status: ticket.StatusInProgress}
	if err := db1.CreateTicket(parent, "tester"); err != nil { // create
		t.Fatal(err)
	}
	a := &ticket.Ticket{Title: "A", NodeType: ticket.NodeLeaf, Status: ticket.StatusTodo, WorkType: "technical_implementation", ParentID: parent.ID}
	b := &ticket.Ticket{Title: "B", NodeType: ticket.NodeLeaf, Status: ticket.StatusTodo, WorkType: "technical_implementation", ParentID: parent.ID}
	if err := db1.CreateTicket(a, "tester"); err != nil {
		t.Fatal(err)
	}
	if err := db1.CreateTicket(b, "tester"); err != nil {
		t.Fatal(err)
	}
	if err := db1.TransitionTicket(a.ID, ticket.StatusInProgress, "tester"); err != nil { // transition
		t.Fatal(err)
	}
	if err := db1.AddDependency(a.ID, b.ID, "tester"); err != nil { // dependency
		t.Fatal(err)
	}
	if _, err := db1.AppendDecision(decision.Record{ // decision record / blocker
		EventType: decision.EventInputRequested, TicketID: a.ID, Status: decision.StatusPending,
		Statement: "Which timeout?",
	}); err != nil {
		t.Fatal(err)
	}

	// Sidecars exist on disk (write-through happened, not a deferred export).
	for _, id := range []string{parent.ID, a.ID, b.ID} {
		if _, err := os.Stat(filepath.Join(ticketsDir, id, "ticket.md")); err != nil {
			t.Fatalf("ticket.md missing for %s: %v", id, err)
		}
	}
	if _, err := os.Stat(decision.Path(ticketsDir, a.ID)); err != nil {
		t.Fatalf("decisions.ndjson missing for %s: %v", a.ID, err)
	}
	db1.Close()

	// Purge: rebuild a fresh store from files only.
	db2, err := sqlite.Open(filepath.Join(t.TempDir(), "fresh.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	defer db2.Close()
	if err := db2.Migrate(); err != nil {
		t.Fatal(err)
	}
	if _, err := importExports(db2, ticketsDir); err != nil {
		t.Fatalf("rebuild import: %v", err)
	}

	// Ticket records + statuses preserved.
	gotA, err := db2.GetTicket(a.ID)
	if err != nil || gotA.Status != ticket.StatusInProgress {
		t.Fatalf("A status = %v err=%v, want in_progress", gotA.Status, err)
	}
	// Dependency preserved.
	deps, err := db2.DependencyIDs(a.ID)
	if err != nil || len(deps) != 1 || deps[0] != b.ID {
		t.Fatalf("A deps = %v err=%v, want [%s]", deps, err, b.ID)
	}
	// Decision record preserved.
	recs, err := db2.ListDecisions(a.ID)
	if err != nil || len(recs) != 1 || recs[0].EventType != decision.EventInputRequested {
		t.Fatalf("A decisions = %+v err=%v", recs, err)
	}
}
