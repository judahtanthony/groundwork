package sqlite

import (
	"database/sql"
	"errors"

	"groundwork/internal/encoding"
)

// Dependency-edge errors.
var (
	// ErrSelfDependency is returned when a node would depend on itself.
	ErrSelfDependency = errors.New("a node cannot depend on itself")
	// ErrDependencyCycle is returned when an edge would introduce a cycle.
	ErrDependencyCycle = errors.New("dependency edge would create a cycle")
)

// DependencyIDs returns the ids that fromID depends on (its out-edges in the
// dependency overlay), sorted for deterministic output.
func (db *DB) DependencyIDs(fromID string) ([]string, error) {
	return queryIDs(db, `SELECT to_id FROM dependencies WHERE from_id=? ORDER BY to_id`, fromID)
}

// DependentIDs returns the ids that depend on toID (its in-edges), sorted.
func (db *DB) DependentIDs(toID string) ([]string, error) {
	return queryIDs(db, `SELECT from_id FROM dependencies WHERE to_id=? ORDER BY from_id`, toID)
}

// AddDependency records that fromID depends on toID. It rejects self-edges and
// any edge that would create a cycle (ADR 0010), and is idempotent if the edge
// already exists. Both nodes must exist. An audit event is appended.
func (db *DB) AddDependency(fromID, toID, actor string) error {
	if fromID == toID {
		return ErrSelfDependency
	}
	return db.withTx(func(tx *sql.Tx) error {
		if err := mustExist(tx, fromID); err != nil {
			return err
		}
		if err := mustExist(tx, toID); err != nil {
			return err
		}

		// A cycle would result if toID can already reach fromID by following
		// depends-on edges (then fromID->toID closes the loop).
		reaches, err := pathExistsTx(tx, toID, fromID)
		if err != nil {
			return err
		}
		if reaches {
			return ErrDependencyCycle
		}

		res, err := tx.Exec(
			`INSERT OR IGNORE INTO dependencies (from_id, to_id, created_at) VALUES (?,?,?)`,
			fromID, toID, encoding.Now(),
		)
		if err != nil {
			return err
		}
		if n, _ := res.RowsAffected(); n == 0 {
			return nil // edge already present: idempotent no-op, no audit
		}
		return appendAudit(tx, actor, "dependency.added", "ticket", fromID, map[string]any{
			"depends_on": toID,
		})
	})
}

// RemoveDependency deletes the edge fromID -> toID if present, appending an
// audit event when a row is removed.
func (db *DB) RemoveDependency(fromID, toID, actor string) error {
	return db.withTx(func(tx *sql.Tx) error {
		res, err := tx.Exec(`DELETE FROM dependencies WHERE from_id=? AND to_id=?`, fromID, toID)
		if err != nil {
			return err
		}
		if n, _ := res.RowsAffected(); n == 0 {
			return ErrNotFound
		}
		return appendAudit(tx, actor, "dependency.removed", "ticket", fromID, map[string]any{
			"depends_on": toID,
		})
	})
}

// --- helpers ---

func queryIDs(db *DB, query string, arg string) ([]string, error) {
	rows, err := db.Query(query, arg)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	ids := []string{}
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

func mustExist(tx *sql.Tx, id string) error {
	var x string
	err := tx.QueryRow(`SELECT id FROM tickets WHERE id=?`, id).Scan(&x)
	if err == sql.ErrNoRows {
		return ErrNotFound
	}
	return err
}

// pathExistsTx reports whether target is reachable from start by following
// depends-on (from_id -> to_id) edges. Iterative DFS with a visited set.
func pathExistsTx(tx *sql.Tx, start, target string) (bool, error) {
	visited := map[string]bool{}
	stack := []string{start}
	for len(stack) > 0 {
		cur := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if cur == target {
			return true, nil
		}
		if visited[cur] {
			continue
		}
		visited[cur] = true

		rows, err := tx.Query(`SELECT to_id FROM dependencies WHERE from_id=?`, cur)
		if err != nil {
			return false, err
		}
		var next []string
		for rows.Next() {
			var id string
			if err := rows.Scan(&id); err != nil {
				rows.Close()
				return false, err
			}
			next = append(next, id)
		}
		rows.Close()
		if err := rows.Err(); err != nil {
			return false, err
		}
		stack = append(stack, next...)
	}
	return false, nil
}
