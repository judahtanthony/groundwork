package sqlite

import (
	"testing"

	"groundwork/internal/run"
	"groundwork/internal/ticket"
)

// insertRun adds a run row directly (bypassing the claim transaction) so a node
// can carry multiple historical runs in a test.
func insertRun(t *testing.T, db *DB, runID, ticketID, startedAt string) {
	t.Helper()
	_, err := db.Exec(`INSERT INTO runs
		(id, ticket_id, actor_id, actor_snapshot_json, mode, runtime, model, status,
		 workspace_path, started_at, updated_at)
		VALUES (?,?,?,?,?,?,?,?,?,?,?)`,
		runID, ticketID, "ai.codex.default", "{}", string(run.ModeImplementation), "codex", "m",
		string(run.StatusCompleted), "", startedAt, startedAt)
	if err != nil {
		t.Fatal(err)
	}
}

func TestRunChangedFilesRoundTrip(t *testing.T) {
	db := openTestDB(t)
	id := seedTicketStatus(t, db, "work", ticket.StatusInProgress)
	runID := "R-1"
	insertRun(t, db, runID, id, "2026-06-28T10:00:00Z")

	// Default is an empty set.
	files, err := db.RunChangedFiles(runID)
	if err != nil || len(files) != 0 {
		t.Fatalf("default: files=%v err=%v", files, err)
	}

	if err := db.SetRunChangedFiles(runID, []string{"b.go", "a.go"}); err != nil {
		t.Fatal(err)
	}
	files, err = db.RunChangedFiles(runID)
	if err != nil || len(files) != 2 {
		t.Fatalf("after set: files=%v err=%v", files, err)
	}
}

func TestLatestInterruptedRunForNode(t *testing.T) {
	db := openTestDB(t)
	id := seedTicketStatus(t, db, "work", ticket.StatusInProgress)

	// No runs → none.
	if got, _ := db.LatestInterruptedRunForNode(id); got != "" {
		t.Fatalf("got %q, want empty", got)
	}
	// A completed run is not a resume candidate.
	insertRun(t, db, "R-1", id, "2026-06-28T10:00:00Z") // status completed
	if got, _ := db.LatestInterruptedRunForNode(id); got != "" {
		t.Fatalf("completed run returned: %q", got)
	}
	// An interrupted run is the resume candidate.
	if _, err := db.Exec(`INSERT INTO runs
		(id, ticket_id, actor_id, actor_snapshot_json, mode, runtime, model, status, workspace_path, started_at, updated_at)
		VALUES (?,?,?,?,?,?,?,?,?,?,?)`,
		"R-2", id, "ai.codex.default", "{}", "implementation", "codex", "m", "interrupted", "", "2026-06-28T11:00:00Z", "2026-06-28T11:00:00Z"); err != nil {
		t.Fatal(err)
	}
	if got, _ := db.LatestInterruptedRunForNode(id); got != "R-2" {
		t.Fatalf("got %q, want R-2", got)
	}
}

func TestChangedFilesForNodePrefersLatestNonEmpty(t *testing.T) {
	db := openTestDB(t)
	id := seedTicketStatus(t, db, "work", ticket.StatusInProgress)

	insertRun(t, db, "R-1", id, "2026-06-28T10:00:00Z")
	if err := db.SetRunChangedFiles("R-1", []string{"old.go"}); err != nil {
		t.Fatal(err)
	}
	// A later run with no changes must not mask the earlier diff.
	insertRun(t, db, "R-2", id, "2026-06-28T11:00:00Z")
	if err := db.SetRunChangedFiles("R-2", nil); err != nil {
		t.Fatal(err)
	}

	files, err := db.ChangedFilesForNode(id)
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 1 || files[0] != "old.go" {
		t.Fatalf("ChangedFilesForNode = %v, want [old.go]", files)
	}
}
