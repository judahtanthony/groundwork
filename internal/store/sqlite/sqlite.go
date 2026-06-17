// Package sqlite provides the Groundwork operational store backed by a pure-Go
// SQLite driver (ADR 0017). It owns connection setup, schema migrations
// (ADR 0018), and the typed store methods built on top.
package sqlite

import (
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

// DB wraps *sql.DB with Groundwork-specific setup.
type DB struct {
	*sql.DB
	path string
}

// Open opens (creating if necessary) the SQLite database at path with the
// pragmas required by docs/contracts/sqlite-schema.md: WAL journaling, foreign
// keys on, and a 5s busy timeout. Transactions take the write lock at BEGIN
// (_txlock=immediate) so the claim path (T-0117) is race-free.
func Open(path string) (*DB, error) {
	sdb, err := sql.Open("sqlite", dsn(path))
	if err != nil {
		return nil, fmt.Errorf("opening sqlite %s: %w", path, err)
	}
	if err := sdb.Ping(); err != nil {
		sdb.Close()
		return nil, fmt.Errorf("opening sqlite %s: %w", path, err)
	}
	return &DB{DB: sdb, path: path}, nil
}

// dsn builds the modernc.org/sqlite DSN. Pragmas are applied on every pooled
// connection via the _pragma query parameters.
func dsn(path string) string {
	return "file:" + path +
		"?_pragma=busy_timeout(5000)" +
		"&_pragma=foreign_keys(on)" +
		"&_pragma=journal_mode(wal)" +
		"&_txlock=immediate"
}

// Path returns the database file path.
func (db *DB) Path() string { return db.path }
