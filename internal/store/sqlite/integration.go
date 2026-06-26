package sqlite

import (
	"database/sql"

	"groundwork/internal/encoding"
)

// IntegrationBranch is a root node's recorded integration target (ADR 0058).
type IntegrationBranch struct {
	NodeID     string `json:"node_id"`
	Branch     string `json:"branch"`
	BaseCommit string `json:"base_commit,omitempty"`
	Status     string `json:"status"`
	CreatedAt  string `json:"created_at"`
	UpdatedAt  string `json:"updated_at"`
}

// RecordIntegrationBranch records (or updates) a root's integration target.
func (db *DB) RecordIntegrationBranch(nodeID, branch, baseCommit string) error {
	now := encoding.Now()
	_, err := db.Exec(`INSERT INTO integration_branches
		(node_id, branch, base_commit, status, created_at, updated_at)
		VALUES (?, ?, ?, 'open', ?, ?)
		ON CONFLICT(node_id) DO UPDATE SET
			branch=excluded.branch, base_commit=excluded.base_commit, updated_at=excluded.updated_at`,
		nodeID, branch, baseCommit, now, now)
	return err
}

// GetIntegrationBranch returns a node's integration target, or (nil, nil) when
// none is recorded.
func (db *DB) GetIntegrationBranch(nodeID string) (*IntegrationBranch, error) {
	var b IntegrationBranch
	err := db.QueryRow(`SELECT node_id, branch, base_commit, status, created_at, updated_at
		FROM integration_branches WHERE node_id = ?`, nodeID).
		Scan(&b.NodeID, &b.Branch, &b.BaseCommit, &b.Status, &b.CreatedAt, &b.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &b, nil
}

// CloseIntegrationBranch marks a root's integration target landed (after the
// gated root land_to_main merge).
func (db *DB) CloseIntegrationBranch(nodeID string) error {
	_, err := db.Exec(`UPDATE integration_branches SET status='landed', updated_at=? WHERE node_id=?`,
		encoding.Now(), nodeID)
	return err
}
