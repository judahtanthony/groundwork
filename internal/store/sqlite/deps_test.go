package sqlite

import (
	"errors"
	"testing"

	"groundwork/internal/ticket"
)

func makeNodes(t *testing.T, db *DB, n int) []string {
	t.Helper()
	ids := make([]string, n)
	for i := 0; i < n; i++ {
		tk := &ticket.Ticket{Title: "node"}
		if err := db.CreateTicket(tk, "human"); err != nil {
			t.Fatal(err)
		}
		ids[i] = tk.ID
	}
	return ids
}

func TestAddAndQueryDependency(t *testing.T) {
	db := openTestDB(t)
	n := makeNodes(t, db, 2)

	if err := db.AddDependency(n[0], n[1], "human"); err != nil {
		t.Fatalf("AddDependency: %v", err)
	}
	deps, _ := db.DependencyIDs(n[0])
	if len(deps) != 1 || deps[0] != n[1] {
		t.Fatalf("DependencyIDs = %v, want [%s]", deps, n[1])
	}
	dependents, _ := db.DependentIDs(n[1])
	if len(dependents) != 1 || dependents[0] != n[0] {
		t.Fatalf("DependentIDs = %v, want [%s]", dependents, n[0])
	}
}

func TestAddDependencyIsIdempotent(t *testing.T) {
	db := openTestDB(t)
	n := makeNodes(t, db, 2)
	if err := db.AddDependency(n[0], n[1], "human"); err != nil {
		t.Fatal(err)
	}
	if err := db.AddDependency(n[0], n[1], "human"); err != nil {
		t.Fatalf("second AddDependency should be a no-op, got %v", err)
	}
	deps, _ := db.DependencyIDs(n[0])
	if len(deps) != 1 {
		t.Fatalf("duplicate edge created: %v", deps)
	}
}

func TestSelfDependencyRejected(t *testing.T) {
	db := openTestDB(t)
	n := makeNodes(t, db, 1)
	if err := db.AddDependency(n[0], n[0], "human"); !errors.Is(err, ErrSelfDependency) {
		t.Fatalf("want ErrSelfDependency, got %v", err)
	}
}

func TestCycleRejected(t *testing.T) {
	db := openTestDB(t)
	n := makeNodes(t, db, 3)
	// A->B->C, then C->A must be rejected.
	if err := db.AddDependency(n[0], n[1], "human"); err != nil {
		t.Fatal(err)
	}
	if err := db.AddDependency(n[1], n[2], "human"); err != nil {
		t.Fatal(err)
	}
	if err := db.AddDependency(n[2], n[0], "human"); !errors.Is(err, ErrDependencyCycle) {
		t.Fatalf("want ErrDependencyCycle, got %v", err)
	}
}

func TestRemoveDependency(t *testing.T) {
	db := openTestDB(t)
	n := makeNodes(t, db, 2)
	if err := db.AddDependency(n[0], n[1], "human"); err != nil {
		t.Fatal(err)
	}
	if err := db.RemoveDependency(n[0], n[1], "human"); err != nil {
		t.Fatal(err)
	}
	deps, _ := db.DependencyIDs(n[0])
	if len(deps) != 0 {
		t.Fatalf("edge not removed: %v", deps)
	}
}
