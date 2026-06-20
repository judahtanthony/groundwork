package sqlite

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"groundwork/internal/encoding"
	"groundwork/internal/ticket"
)

// ErrValidationGate is returned when landing is blocked because required
// validation has not passed (and no override was given).
var ErrValidationGate = errors.New("required validation has not passed")

// ErrNotApproved is returned when landing a node that is not ready to land
// (it must be in review or approved; the human landing gate is the approval).
var ErrNotApproved = errors.New("node must be in review or approved before landing")

// LandResult reports what blocked a landing, for actionable errors.
type LandResult struct {
	Missing []string // required checks with no passing result
	Failed  []string // checks whose latest result failed
}

// Land lands an approved node: it enforces the validation gate (every required
// check must have a passing result and none may be failing, unless override),
// then transitions approved -> landing -> done and squashes the latest run's WIP
// checkpoints (ADR 0015). requiredChecks come from the validation policy applied
// to the node's changed files (the file set is supplied by the Phase 4 runtime;
// in M2 it is typically empty, so the gate enforces "no failing results").
func (db *DB) Land(ticketID string, requiredChecks []string, override bool, actor string) (*LandResult, error) {
	results, err := db.ListValidationsForTicket(ticketID)
	if err != nil {
		return nil, err
	}
	passed := map[string]bool{}
	var failed []string
	for _, r := range results {
		switch r.Status {
		case ValidationPass:
			passed[r.Name] = true
		case ValidationFail:
			failed = append(failed, r.Name)
		}
	}
	var missing []string
	for _, name := range requiredChecks {
		if !passed[name] {
			missing = append(missing, name)
		}
	}
	if (len(missing) > 0 || len(failed) > 0) && !override {
		return &LandResult{Missing: missing, Failed: failed},
			fmt.Errorf("%w: missing=%v failed=%v", ErrValidationGate, missing, failed)
	}

	var latestRun string
	err = db.withTx(func(tx *sql.Tx) error {
		var status string
		if err := tx.QueryRow(`SELECT status FROM tickets WHERE id=?`, ticketID).Scan(&status); err != nil {
			if err == sql.ErrNoRows {
				return ErrNotFound
			}
			return err
		}
		// The land_to_main approval is the human gate; a node ready to land sits in
		// review (prepared work) or approved (proposal accepted). Both are valid
		// landing origins.
		if st := ticket.Status(status); st != ticket.StatusReview && st != ticket.StatusApproved {
			return ErrNotApproved
		}
		now := encoding.Now()
		for _, to := range []ticket.Status{ticket.StatusLanding, ticket.StatusDone} {
			if _, err := tx.Exec(`UPDATE tickets SET status=?, updated_at=? WHERE id=?`, string(to), now, ticketID); err != nil {
				return err
			}
		}
		_ = tx.QueryRow(`SELECT id FROM runs WHERE ticket_id=? ORDER BY id DESC LIMIT 1`, ticketID).Scan(&latestRun)
		payload := map[string]any{"to": ticket.StatusDone}
		if override {
			payload["validation_override"] = true
		}
		return appendAudit(tx, actor, "ticket.landed", "ticket", ticketID, payload)
	})
	if err != nil {
		return nil, err
	}
	// Squash WIP checkpoints into the landing commit (records-only in M2; the git
	// squash is Phase 4). Done outside the landing tx since it opens its own.
	if latestRun != "" {
		_, _ = db.SquashCheckpoints(latestRun)
	}
	return nil, nil
}

// LandResultError renders a LandResult for display.
func (lr *LandResult) String() string {
	var b strings.Builder
	if len(lr.Missing) > 0 {
		fmt.Fprintf(&b, "missing required validation: %s. ", strings.Join(lr.Missing, ", "))
	}
	if len(lr.Failed) > 0 {
		fmt.Fprintf(&b, "failed validation: %s.", strings.Join(lr.Failed, ", "))
	}
	return b.String()
}
