package sqlite

import (
	"database/sql"
	"errors"
	"time"

	"groundwork/internal/encoding"
	"groundwork/internal/ticket"
)

// Lease-related errors.
var (
	// ErrNotEligible is returned when claiming a node that is not dispatchable.
	ErrNotEligible = errors.New("node is not eligible for claim")
	// ErrAlreadyLeased is returned when a node already has an active lease.
	ErrAlreadyLeased = errors.New("node is already leased")
	// ErrLeaseNotHeld is returned when renewing/releasing a lease not held by
	// the given run.
	ErrLeaseNotHeld = errors.New("lease not held by this run")
	// ErrLeaseExpired is returned when renewing a lease that has already expired;
	// it must be re-claimed rather than renewed.
	ErrLeaseExpired = errors.New("lease has expired and cannot be renewed")
)

const leaseActiveStatus = "active"

// Lease is an exclusive active-work claim on a node.
type Lease struct {
	TicketID  string `json:"ticket_id"`
	RunID     string `json:"run_id"`
	ActorID   string `json:"actor_id"`
	Status    string `json:"status"`
	ExpiresAt string `json:"expires_at"`
	RenewedAt string `json:"renewed_at"`
}

// ClaimTicket atomically claims an eligible node: it verifies eligibility
// (todo + dependencies satisfied), ensures no active lease exists, records a
// lease, and moves the node to in_progress — all in one transaction so only one
// run can win. Appends a ticket.claimed audit event.
//
// A stale (expired) lease on an otherwise-eligible node is cleared and the node
// reclaimed. Note that in Phase 1 nothing returns an in_progress node to todo,
// so this branch is reached only once the Phase 2 recovery sweep
// (docs/architecture/recovery.md) requeues interrupted runs.
func (db *DB) ClaimTicket(ticketID, runID, actorID string, ttl time.Duration) (*Lease, error) {
	now := time.Now()
	var lease *Lease
	err := db.withTx(func(tx *sql.Tx) error {
		l, err := txClaim(tx, ticketID, runID, actorID, now, ttl)
		lease = l
		return err
	})
	if err != nil {
		return nil, err
	}
	return lease, nil
}

// txClaim is the transactional single-winner claim core shared by ClaimTicket
// and StartRun: it verifies eligibility (todo + dependencies satisfied), rejects
// an active lease (clearing an expired one), records the lease, moves the node to
// in_progress, and appends a ticket.claimed audit event. It must run inside a
// write transaction so only one run can win.
func txClaim(tx *sql.Tx, ticketID, runID, actorID string, now time.Time, ttl time.Duration) (*Lease, error) {
	var statusStr string
	err := tx.QueryRow(`SELECT status FROM tickets WHERE id=?`, ticketID).Scan(&statusStr)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	if ticket.Status(statusStr) != ticket.StatusTodo {
		return nil, ErrNotEligible
	}
	satisfied, err := txDependenciesSatisfied(tx, ticketID)
	if err != nil {
		return nil, err
	}
	if !satisfied {
		return nil, ErrNotEligible
	}

	// Reject if an unexpired lease exists; clear an expired one.
	var expiresAt string
	err = tx.QueryRow(`SELECT expires_at FROM leases WHERE ticket_id=?`, ticketID).Scan(&expiresAt)
	switch {
	case err == nil:
		if leaseIsActive(expiresAt, now) {
			return nil, ErrAlreadyLeased
		}
		if _, err := tx.Exec(`DELETE FROM leases WHERE ticket_id=?`, ticketID); err != nil {
			return nil, err
		}
	case err == sql.ErrNoRows:
		// no existing lease
	default:
		return nil, err
	}

	nowStr := encoding.FormatTime(now)
	expStr := encoding.FormatTime(now.Add(ttl))
	if _, err := tx.Exec(
		`INSERT INTO leases (ticket_id, run_id, actor_id, status, expires_at, renewed_at)
		 VALUES (?,?,?,?,?,?)`,
		ticketID, runID, actorID, leaseActiveStatus, expStr, nowStr,
	); err != nil {
		return nil, err
	}
	if _, err := tx.Exec(`UPDATE tickets SET status=?, updated_at=? WHERE id=?`,
		string(ticket.StatusInProgress), nowStr, ticketID); err != nil {
		return nil, err
	}
	if err := appendAudit(tx, actorID, "ticket.claimed", "ticket", ticketID, map[string]any{
		"run_id": runID,
	}); err != nil {
		return nil, err
	}
	return &Lease{
		TicketID: ticketID, RunID: runID, ActorID: actorID,
		Status: leaseActiveStatus, ExpiresAt: expStr, RenewedAt: nowStr,
	}, nil
}

// RenewLease extends a lease held by runID. It fails if the lease is missing,
// held by a different run, or already expired (an expired lease must be
// re-claimed, not renewed). It appends a ticket.lease_renewed audit event.
func (db *DB) RenewLease(ticketID, runID string, ttl time.Duration) (*Lease, error) {
	now := time.Now()
	var lease *Lease
	err := db.withTx(func(tx *sql.Tx) error {
		var holder, actorID, expiresAt string
		err := tx.QueryRow(
			`SELECT run_id, actor_id, expires_at FROM leases WHERE ticket_id=?`, ticketID,
		).Scan(&holder, &actorID, &expiresAt)
		if err == sql.ErrNoRows {
			return ErrLeaseNotHeld
		}
		if err != nil {
			return err
		}
		if holder != runID {
			return ErrLeaseNotHeld
		}
		if !leaseIsActive(expiresAt, now) {
			return ErrLeaseExpired
		}
		nowStr := encoding.FormatTime(now)
		expStr := encoding.FormatTime(now.Add(ttl))
		if _, err := tx.Exec(
			`UPDATE leases SET expires_at=?, renewed_at=? WHERE ticket_id=?`,
			expStr, nowStr, ticketID,
		); err != nil {
			return err
		}
		if err := appendAudit(tx, actorID, "ticket.lease_renewed", "ticket", ticketID, map[string]any{
			"run_id": runID,
		}); err != nil {
			return err
		}
		lease = &Lease{TicketID: ticketID, RunID: runID, ActorID: actorID, Status: leaseActiveStatus, ExpiresAt: expStr, RenewedAt: nowStr}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return lease, nil
}

// ReleaseLease removes a lease held by runID, appending a ticket.lease_released
// audit event. It does not change the node's status.
func (db *DB) ReleaseLease(ticketID, runID string) error {
	return db.withTx(func(tx *sql.Tx) error {
		var holder string
		err := tx.QueryRow(`SELECT run_id FROM leases WHERE ticket_id=?`, ticketID).Scan(&holder)
		if err == sql.ErrNoRows {
			return ErrLeaseNotHeld
		}
		if err != nil {
			return err
		}
		if holder != runID {
			return ErrLeaseNotHeld
		}
		if _, err := tx.Exec(`DELETE FROM leases WHERE ticket_id=?`, ticketID); err != nil {
			return err
		}
		return appendAudit(tx, runID, "ticket.lease_released", "ticket", ticketID, nil)
	})
}

// GetLease returns the lease for a node, or (nil, nil) if none exists.
func (db *DB) GetLease(ticketID string) (*Lease, error) {
	var l Lease
	err := db.QueryRow(
		`SELECT ticket_id, run_id, actor_id, status, expires_at, renewed_at FROM leases WHERE ticket_id=?`,
		ticketID,
	).Scan(&l.TicketID, &l.RunID, &l.ActorID, &l.Status, &l.ExpiresAt, &l.RenewedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &l, nil
}

// txDependenciesSatisfied reports whether all of fromID's dependencies are done,
// within the given transaction.
func txDependenciesSatisfied(tx *sql.Tx, fromID string) (bool, error) {
	rows, err := tx.Query(
		`SELECT t.status FROM dependencies d JOIN tickets t ON t.id = d.to_id WHERE d.from_id=?`,
		fromID,
	)
	if err != nil {
		return false, err
	}
	defer rows.Close()
	for rows.Next() {
		var s string
		if err := rows.Scan(&s); err != nil {
			return false, err
		}
		if !ticket.DependencyMet(ticket.Status(s)) {
			return false, nil
		}
	}
	return true, rows.Err()
}

// leaseIsActive reports whether a lease with the given expiry is still active.
func leaseIsActive(expiresAt string, now time.Time) bool {
	exp, err := encoding.ParseTime(expiresAt)
	if err != nil {
		return false
	}
	return exp.After(now)
}
