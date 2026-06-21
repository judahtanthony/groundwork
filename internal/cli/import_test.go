package cli

import (
	"os"
	"path/filepath"
	"testing"

	"groundwork/internal/exporter"
	"groundwork/internal/store/sqlite"
	"groundwork/internal/ticket"
)

// writeExport renders tk to dir/<id>/ticket.md the same way `gw ticket export`
// does, so importExports parses a faithful fixture.
func writeExport(t *testing.T, dir string, tk *ticket.Ticket, deps []string) {
	t.Helper()
	data, err := exporter.Render(tk, deps)
	if err != nil {
		t.Fatalf("render %s: %v", tk.ID, err)
	}
	d := filepath.Join(dir, tk.ID)
	if err := os.MkdirAll(d, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(d, "ticket.md"), data, 0o644); err != nil {
		t.Fatal(err)
	}
}

// TestImportWholeTreePreservesHierarchyStatusAndSeq exercises the whole-tree
// import path (T-1006, ADR 0032): goal/epic non-T parent ids, a completed node,
// a cross-node dependency, and the allocator reseeding from the max *T* id only.
func TestImportWholeTreePreservesHierarchyStatusAndSeq(t *testing.T) {
	dir := t.TempDir()
	// G-2000 carries a deliberately high numeric suffix: it must NOT feed the
	// T-id allocator, or fresh tickets would skip to T-2001.
	goal := &ticket.Ticket{ID: "G-2000", Kind: "goal", Title: "Bootstrap goal", Status: ticket.StatusDone}
	epic := &ticket.Ticket{ID: "E-0011", Kind: "epic", Title: "Dogfood epic", Status: ticket.StatusTodo, ParentID: "G-2000"}
	done := &ticket.Ticket{ID: "T-1003", Kind: "ticket", Title: "Completed ticket", Status: ticket.StatusDone, ParentID: "E-0011"}
	open := &ticket.Ticket{ID: "T-1002", Kind: "ticket", Title: "Open ticket", Status: ticket.StatusTodo, ParentID: "E-0011"}
	writeExport(t, dir, goal, nil)
	writeExport(t, dir, epic, nil)
	writeExport(t, dir, done, nil)
	writeExport(t, dir, open, []string{"T-1003"})

	db, err := sqlite.Open(filepath.Join(t.TempDir(), "state.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err := db.Migrate(); err != nil {
		t.Fatal(err)
	}

	n, err := importExports(db, dir)
	if err != nil {
		t.Fatalf("import: %v", err)
	}
	if n != 4 {
		t.Fatalf("imported %d nodes, want 4", n)
	}

	// Hierarchy preserved across non-T parent ids.
	gotEpic, err := db.GetTicket("E-0011")
	if err != nil || gotEpic.ParentID != "G-2000" {
		t.Fatalf("epic parent = %q (err %v), want G-2000", gotEpic.ParentID, err)
	}
	// Completed status preserved.
	gotDone, err := db.GetTicket("T-1003")
	if err != nil || gotDone.Status != ticket.StatusDone {
		t.Fatalf("T-1003 status = %q (err %v), want done", gotDone.Status, err)
	}
	// Dependency edge rebuilt.
	deps, err := db.DependencyIDs("T-1002")
	if err != nil || len(deps) != 1 || deps[0] != "T-1003" {
		t.Fatalf("T-1002 deps = %v (err %v), want [T-1003]", deps, err)
	}

	// The allocator reseeds from the max T id (1003), not the goal's 2000.
	fresh := &ticket.Ticket{Title: "freshly created", Status: ticket.StatusBacklog}
	if err := db.CreateTicket(fresh, "human.owner"); err != nil {
		t.Fatal(err)
	}
	if fresh.ID != "T-1004" {
		t.Fatalf("next allocated id = %q, want T-1004", fresh.ID)
	}
}
