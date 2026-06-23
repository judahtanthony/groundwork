package sqlite

import (
	"bytes"
	"testing"

	"groundwork/internal/exporter"
	"groundwork/internal/ticket"
)

// renderNode exports a node the way `gw ticket export` does: the stored ticket
// plus its (deterministically ordered) dependency ids.
func renderNode(t *testing.T, db *DB, id string) []byte {
	t.Helper()
	tk, err := db.GetTicket(id)
	if err != nil {
		t.Fatalf("get %s: %v", id, err)
	}
	deps, err := db.DependencyIDs(id)
	if err != nil {
		t.Fatalf("deps %s: %v", id, err)
	}
	out, err := exporter.Render(tk, deps)
	if err != nil {
		t.Fatalf("render %s: %v", id, err)
	}
	return out
}

// TestImportExportRoundTripIsByteStable locks the byte-stability guarantee a cold
// rebuild relies on (ADR 0021, T-1032): a store rebuilt from committed exports
// must re-export byte-for-byte identically. It exercises the two properties that
// guarantee it — empty timestamps survive ImportTicket, and dependencies export
// in deterministic id order regardless of insertion order.
func TestImportExportRoundTripIsByteStable(t *testing.T) {
	ids := []string{"T-0001", "T-0002", "T-0003"}

	build := func() *DB {
		db := openTestDB(t)
		for _, n := range []*ticket.Ticket{
			{ID: "T-0001", Kind: "ticket", Title: "a", Status: ticket.StatusDone},
			{ID: "T-0002", Kind: "ticket", Title: "b", Status: ticket.StatusDone},
			{ID: "T-0003", Kind: "ticket", Title: "c", Status: ticket.StatusTodo},
		} {
			if err := db.ImportTicket(n); err != nil {
				t.Fatalf("import %s: %v", n.ID, err)
			}
		}
		// Add dependencies out of id order: export must still emit them sorted.
		if err := db.AddDependency("T-0003", "T-0002", "t"); err != nil {
			t.Fatal(err)
		}
		if err := db.AddDependency("T-0003", "T-0001", "t"); err != nil {
			t.Fatal(err)
		}
		return db
	}

	db1 := build()
	canonical := map[string][]byte{}
	for _, id := range ids {
		canonical[id] = renderNode(t, db1, id)
	}

	// Empty timestamps survive the import (no synthesized "now").
	c3 := canonical["T-0003"]
	if !bytes.Contains(c3, []byte(`created_at: ""`)) || !bytes.Contains(c3, []byte(`updated_at: ""`)) {
		t.Fatalf("export did not preserve empty timestamps:\n%s", c3)
	}
	// Dependencies are id-sorted regardless of the insertion order above.
	if i, j := bytes.Index(c3, []byte("T-0001")), bytes.Index(c3, []byte("T-0002")); i < 0 || j < 0 || i > j {
		t.Fatalf("dependencies not in deterministic id order:\n%s", c3)
	}

	// Rebuild a fresh store from the canonical exports (parse -> import -> link),
	// then re-export: every node must be byte-identical to the canonical form.
	db2 := openTestDB(t)
	parsedDeps := map[string][]string{}
	for _, id := range ids {
		tk, deps, err := exporter.Parse(canonical[id])
		if err != nil {
			t.Fatalf("parse %s: %v", id, err)
		}
		if err := db2.ImportTicket(tk); err != nil {
			t.Fatalf("reimport %s: %v", id, err)
		}
		parsedDeps[id] = deps
	}
	for _, id := range ids {
		for _, dep := range parsedDeps[id] {
			if err := db2.AddDependency(id, dep, "t"); err != nil {
				t.Fatalf("relink %s->%s: %v", id, dep, err)
			}
		}
	}
	for _, id := range ids {
		if got := renderNode(t, db2, id); !bytes.Equal(got, canonical[id]) {
			t.Fatalf("%s not byte-stable across rebuild:\ncanonical:\n%s\ngot:\n%s", id, canonical[id], got)
		}
	}
}
