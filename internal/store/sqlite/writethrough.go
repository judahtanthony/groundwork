package sqlite

import (
	"errors"
	"fmt"

	"groundwork/internal/decision"
	"groundwork/internal/exporter"
)

// SetExportDir enables filesystem write-through to the given tickets directory
// (ADR 0053). After this is set, durable ticket/dependency/decision mutations
// rewrite the affected node's sidecar (ticket.md + decisions.ndjson) so the
// filesystem source of truth is updated before the mutation reports success.
// Passing "" disables write-through (store-only).
func (db *DB) SetExportDir(dir string) { db.exportDir = dir }

// ExportDir reports the configured write-through directory ("" when disabled).
func (db *DB) ExportDir() string { return db.exportDir }

// writeThrough rewrites the sidecars for the given node ids when write-through is
// enabled. It is called by durable mutation methods after their transaction
// commits. A failure is returned to the caller: SQLite has committed, but the
// caller learns the filesystem is not yet consistent, and startup divergence
// detection (DetectFileDivergence) is the backstop.
func (db *DB) writeThrough(ids ...string) error {
	if db.exportDir == "" {
		return nil
	}
	for _, id := range ids {
		if err := db.exportNode(id); err != nil {
			return fmt.Errorf("durable export %s: %w", id, err)
		}
	}
	return nil
}

// exportNode writes a node's authoritative sidecars from current store state:
// the ticket.md export (with its dependency ids) and the decisions.ndjson
// sidecar. A node that no longer exists is a no-op.
func (db *DB) exportNode(id string) error {
	t, err := db.GetTicket(id)
	if errors.Is(err, ErrNotFound) {
		return nil
	}
	if err != nil {
		return err
	}
	deps, err := db.DependencyIDs(id)
	if err != nil {
		return err
	}
	if _, err := exporter.WriteTo(db.exportDir, t, deps); err != nil {
		return err
	}
	recs, err := db.ListDecisions(id)
	if err != nil {
		return err
	}
	return decision.Write(db.exportDir, id, recs)
}
