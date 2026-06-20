package sqlite

import (
	"path/filepath"
	"testing"
)

// openTestDB opens a migrated database in a temp dir.
func openTestDB(t *testing.T) *DB {
	t.Helper()
	db, err := Open(filepath.Join(t.TempDir(), "state.sqlite"))
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	if err := db.Migrate(); err != nil {
		t.Fatalf("Migrate: %v", err)
	}
	return db
}

func TestOpenAppliesPragmas(t *testing.T) {
	db, err := Open(filepath.Join(t.TempDir(), "state.sqlite"))
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer db.Close()

	var journal string
	if err := db.QueryRow(`PRAGMA journal_mode`).Scan(&journal); err != nil {
		t.Fatal(err)
	}
	if journal != "wal" {
		t.Errorf("journal_mode = %q, want wal", journal)
	}

	var fk int
	if err := db.QueryRow(`PRAGMA foreign_keys`).Scan(&fk); err != nil {
		t.Fatal(err)
	}
	if fk != 1 {
		t.Errorf("foreign_keys = %d, want 1", fk)
	}

	var busy int
	if err := db.QueryRow(`PRAGMA busy_timeout`).Scan(&busy); err != nil {
		t.Fatal(err)
	}
	if busy != 5000 {
		t.Errorf("busy_timeout = %d, want 5000", busy)
	}
}

func TestMigrateCreatesSchema(t *testing.T) {
	db := openTestDB(t)
	for _, table := range []string{
		"meta", "tickets", "dependencies", "leases", "audit_events", "schema_migrations",
		"runs", "run_events", "approvals", "validation_results",
	} {
		var name string
		err := db.QueryRow(`SELECT name FROM sqlite_master WHERE type='table' AND name=?`, table).Scan(&name)
		if err != nil {
			t.Errorf("table %s missing: %v", table, err)
		}
	}
}

func TestMigrateIsIdempotent(t *testing.T) {
	db := openTestDB(t)
	// Second run must be a no-op (no error, no duplicate records).
	if err := db.Migrate(); err != nil {
		t.Fatalf("second Migrate: %v", err)
	}
	// Two runs must record each embedded migration exactly once.
	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM schema_migrations`).Scan(&count); err != nil {
		t.Fatal(err)
	}
	embedded, err := loadMigrations()
	if err != nil {
		t.Fatal(err)
	}
	if count != len(embedded) {
		t.Errorf("schema_migrations rows = %d, want %d", count, len(embedded))
	}
}

func TestMigrateDetectsChecksumDrift(t *testing.T) {
	db := openTestDB(t)
	// Corrupt the recorded checksum to simulate an edited applied migration.
	if _, err := db.Exec(`UPDATE schema_migrations SET checksum = 'tampered'`); err != nil {
		t.Fatal(err)
	}
	if err := db.Migrate(); err == nil {
		t.Fatal("expected checksum drift error, got nil")
	}
}

func TestForeignKeysEnforced(t *testing.T) {
	db := openTestDB(t)
	// Inserting a dependency referencing a missing ticket must fail with FK on.
	_, err := db.Exec(`INSERT INTO dependencies (from_id, to_id, created_at) VALUES ('T-0001','T-0002','x')`)
	if err == nil {
		t.Fatal("expected foreign key violation, got nil")
	}
}
