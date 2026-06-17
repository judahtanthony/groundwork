package sqlite

import (
	"crypto/sha256"
	"embed"
	"encoding/hex"
	"fmt"
	"io/fs"
	"sort"
	"strings"

	"groundwork/internal/encoding"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

// migration is one embedded SQL file.
type migration struct {
	id   string // file name, e.g. "0001_init.sql"
	sql  string
	hash string // sha256 of sql, hex
}

// Migrate applies all pending embedded migrations in lexical order. It is safe
// to re-run: already-applied migrations are skipped. A migration whose recorded
// checksum no longer matches its file is a hard error (drift detection).
// Migrations are forward-only (ADR 0018).
func (db *DB) Migrate() error {
	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (
		id         TEXT PRIMARY KEY,
		checksum   TEXT NOT NULL,
		applied_at TEXT NOT NULL
	)`); err != nil {
		return fmt.Errorf("creating schema_migrations: %w", err)
	}

	applied, err := db.appliedMigrations()
	if err != nil {
		return err
	}

	migrations, err := loadMigrations()
	if err != nil {
		return err
	}

	for _, m := range migrations {
		if prevHash, ok := applied[m.id]; ok {
			if prevHash != m.hash {
				return fmt.Errorf("migration %s checksum drift: recorded %s, file %s", m.id, prevHash, m.hash)
			}
			continue
		}
		if err := db.applyMigration(m); err != nil {
			return err
		}
	}
	return nil
}

// AppliedMigrationIDs returns the ids of applied migrations, sorted.
func (db *DB) AppliedMigrationIDs() ([]string, error) {
	applied, err := db.appliedMigrations()
	if err != nil {
		return nil, err
	}
	ids := make([]string, 0, len(applied))
	for id := range applied {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids, nil
}

// appliedMigrations returns id -> checksum for migrations already recorded.
func (db *DB) appliedMigrations() (map[string]string, error) {
	rows, err := db.Query(`SELECT id, checksum FROM schema_migrations`)
	if err != nil {
		return nil, fmt.Errorf("reading schema_migrations: %w", err)
	}
	defer rows.Close()

	out := map[string]string{}
	for rows.Next() {
		var id, checksum string
		if err := rows.Scan(&id, &checksum); err != nil {
			return nil, err
		}
		out[id] = checksum
	}
	return out, rows.Err()
}

// applyMigration runs one migration and records it, in a single transaction.
func (db *DB) applyMigration(m migration) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(m.sql); err != nil {
		return fmt.Errorf("applying migration %s: %w", m.id, err)
	}
	if _, err := tx.Exec(
		`INSERT INTO schema_migrations (id, checksum, applied_at) VALUES (?, ?, ?)`,
		m.id, m.hash, encoding.Now(),
	); err != nil {
		return fmt.Errorf("recording migration %s: %w", m.id, err)
	}
	return tx.Commit()
}

// loadMigrations reads and sorts the embedded migration files.
func loadMigrations() ([]migration, error) {
	entries, err := fs.ReadDir(migrationsFS, "migrations")
	if err != nil {
		return nil, err
	}

	var migrations []migration
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".sql") {
			continue
		}
		data, err := migrationsFS.ReadFile("migrations/" + e.Name())
		if err != nil {
			return nil, err
		}
		sum := sha256.Sum256(data)
		migrations = append(migrations, migration{
			id:   e.Name(),
			sql:  string(data),
			hash: hex.EncodeToString(sum[:]),
		})
	}

	sort.Slice(migrations, func(i, j int) bool { return migrations[i].id < migrations[j].id })
	return migrations, nil
}
