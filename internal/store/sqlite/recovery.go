package sqlite

import (
	"database/sql"

	"groundwork/internal/encoding"
	"groundwork/internal/run"
	"groundwork/internal/ticket"
)

// RecoveryReport summarizes what startup reconciliation changed.
type RecoveryReport struct {
	InterruptedRuns int `json:"interrupted_runs"`
	ReleasedLeases  int `json:"released_leases"`
	RequeuedNodes   int `json:"requeued_nodes"`
}

// ReconcileStartup brings the store to a stable point after a crash or restart
// (recovery.md, ADR 0027). With no live workers at startup, every non-terminal
// run is stale: it is marked interrupted, its lease released, and its node
// requeued (in_progress -> todo) so the scheduler can re-dispatch it. Runs that
// reached a terminal state and completed work are untouched.
func (db *DB) ReconcileStartup() (*RecoveryReport, error) {
	rep := &RecoveryReport{}
	now := encoding.Now()
	err := db.withTx(func(tx *sql.Tx) error {
		// 1. Mark non-terminal runs interrupted.
		live := []string{string(run.StatusPending), string(run.StatusRunning), string(run.StatusPaused)}
		res, err := tx.Exec(
			`UPDATE runs SET status=?, updated_at=? WHERE status IN (?,?,?)`,
			string(run.StatusInterrupted), now, live[0], live[1], live[2])
		if err != nil {
			return err
		}
		if n, _ := res.RowsAffected(); n > 0 {
			rep.InterruptedRuns = int(n)
			if err := appendAudit(tx, "recovery", "runs.interrupted", "system", "startup", map[string]any{
				"count": rep.InterruptedRuns,
			}); err != nil {
				return err
			}
		}

		// 2. Release every lease and requeue its node if still in_progress.
		rows, err := tx.Query(`SELECT ticket_id FROM leases`)
		if err != nil {
			return err
		}
		var ticketIDs []string
		for rows.Next() {
			var id string
			if err := rows.Scan(&id); err != nil {
				rows.Close()
				return err
			}
			ticketIDs = append(ticketIDs, id)
		}
		rows.Close()
		if err := rows.Err(); err != nil {
			return err
		}

		for _, id := range ticketIDs {
			if _, err := tx.Exec(`DELETE FROM leases WHERE ticket_id=?`, id); err != nil {
				return err
			}
			rep.ReleasedLeases++
			var status string
			if err := tx.QueryRow(`SELECT status FROM tickets WHERE id=?`, id).Scan(&status); err != nil {
				if err == sql.ErrNoRows {
					continue
				}
				return err
			}
			if ticket.Status(status) == ticket.StatusInProgress {
				if _, err := tx.Exec(`UPDATE tickets SET status=?, updated_at=? WHERE id=?`,
					string(ticket.StatusTodo), now, id); err != nil {
					return err
				}
				if err := appendAudit(tx, "recovery", "ticket.transitioned", "ticket", id, map[string]any{
					"from": ticket.StatusInProgress, "to": ticket.StatusTodo, "reason": "recovery_requeue",
				}); err != nil {
					return err
				}
				rep.RequeuedNodes++
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return rep, nil
}
