package cli

import (
	"bytes"
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

// TestImportExportRoundTripIsByteStable guards the deterministic-export contract
// (ADR 0021): rebuilding the store from committed exports and re-exporting must
// yield byte-identical Markdown. It exercises both churn sources: the committed
// empty-timestamp convention (import must not stamp "now") and depends_on order
// (export must sort), here supplied deliberately unsorted.
func TestImportExportRoundTripIsByteStable(t *testing.T) {
	dir := t.TempDir()
	// Empty timestamps mirror planning-sourced committed tickets; deps are
	// supplied out of order to prove export sorts them.
	a := &ticket.Ticket{ID: "T-0503", Kind: "ticket", Title: "Dep A", Status: ticket.StatusTodo}
	b := &ticket.Ticket{ID: "T-0504", Kind: "ticket", Title: "Dep B", Status: ticket.StatusTodo}
	d := &ticket.Ticket{ID: "T-1002", Kind: "ticket", Title: "Dep C", Status: ticket.StatusTodo}
	c := &ticket.Ticket{ID: "T-1003", Kind: "ticket", Title: "Dependent", Status: ticket.StatusBacklog}
	writeExport(t, dir, a, nil)
	writeExport(t, dir, b, nil)
	writeExport(t, dir, d, nil)
	writeExport(t, dir, c, []string{"T-1002", "T-0503", "T-0504"})

	// Snapshot the committed bytes before any import touches the store.
	want := map[string][]byte{}
	for _, id := range []string{"T-0503", "T-0504", "T-1002", "T-1003"} {
		data, err := os.ReadFile(filepath.Join(dir, id, "ticket.md"))
		if err != nil {
			t.Fatal(err)
		}
		want[id] = data
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

	// Empty timestamps must survive the round trip rather than being stamped.
	got, err := db.GetTicket("T-1003")
	if err != nil {
		t.Fatal(err)
	}
	if got.CreatedAt != "" || got.UpdatedAt != "" {
		t.Fatalf("import stamped timestamps: created=%q updated=%q, want empty", got.CreatedAt, got.UpdatedAt)
	}

	// Re-export from the rebuilt store and compare byte-for-byte.
	depMap, err := db.DependencyMap()
	if err != nil {
		t.Fatal(err)
	}
	for id, wantData := range want {
		tk, err := db.GetTicket(id)
		if err != nil {
			t.Fatal(err)
		}
		gotData, err := exporter.Render(tk, depMap[id])
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(gotData, wantData) {
			t.Errorf("%s re-export differs from committed bytes:\n--- got ---\n%s\n--- want ---\n%s", id, gotData, wantData)
		}
	}
}

// TestImportToleratesMissingDependencyEndpoint asserts a depends_on target absent
// from the export set is skipped (not a hard error), while the import still
// succeeds — the only benign AddDependency error the importer swallows (CR#7).
func TestImportToleratesMissingDependencyEndpoint(t *testing.T) {
	dir := t.TempDir()
	a := &ticket.Ticket{ID: "T-2001", Kind: "ticket", Title: "A", Status: ticket.StatusTodo}
	writeExport(t, dir, a, []string{"T-9999"}) // T-9999 is not in the export set

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
		t.Fatalf("import should tolerate a missing endpoint, got: %v", err)
	}
	if n != 1 {
		t.Fatalf("imported %d, want 1", n)
	}
	deps, err := db.DependencyIDs("T-2001")
	if err != nil {
		t.Fatal(err)
	}
	if len(deps) != 0 {
		t.Errorf("deps = %v, want none (missing endpoint skipped)", deps)
	}
}
