package sqlite

import (
	"database/sql"
	"fmt"
	"time"

	"groundwork/internal/encoding"
	"groundwork/internal/run"
	"groundwork/internal/ticket"
)

// Run is one supervised node attempt (docs/contracts/sqlite-schema.md, ADR 0027).
type Run struct {
	ID            string `json:"id"`
	TicketID      string `json:"ticket_id"`
	ActorID       string `json:"actor_id"`
	ActorSnapshot string `json:"actor_snapshot,omitempty"`
	Mode          string `json:"mode"`
	Runtime       string `json:"runtime"`
	Model         string `json:"model,omitempty"`
	Status        string `json:"status"`
	WorkspacePath string `json:"workspace_path,omitempty"`
	BaseCommit    string `json:"base_commit,omitempty"`
	StartedAt     string `json:"started_at"`
	UpdatedAt     string `json:"updated_at"`
	CompletedAt   string `json:"completed_at,omitempty"`
	LastEvent     string `json:"last_event,omitempty"`
	LastMessage   string `json:"last_message,omitempty"`
	InputTokens   int    `json:"input_tokens"`
	OutputTokens  int    `json:"output_tokens"`
	TotalTokens   int    `json:"total_tokens"`
}

// StartRunParams configures a new run.
type StartRunParams struct {
	TicketID      string
	ActorID       string
	ActorSnapshot string // JSON snapshot of the selected actor config (ADR 0023)
	Mode          run.Mode
	Runtime       string
	Model         string
	WorkspacePath string
	BaseCommit    string
	TTL           time.Duration
}

const runColumns = `id, ticket_id, actor_id, actor_snapshot_json, mode, runtime, model,
	status, workspace_path, base_commit, started_at, updated_at, completed_at,
	last_event, last_message, input_tokens, output_tokens, total_tokens`

// StartRun atomically claims an eligible node and creates its run record and
// lease in one transaction, so only one run can win a node (ADR 0026). The run
// starts in `running`. Returns the created run and its lease.
func (db *DB) StartRun(p StartRunParams) (*Run, *Lease, error) {
	if !p.Mode.Valid() {
		return nil, nil, fmt.Errorf("invalid run mode %q", p.Mode)
	}
	now := time.Now()
	nowStr := encoding.FormatTime(now)
	var r *Run
	var lease *Lease
	err := db.withTx(func(tx *sql.Tx) error {
		runID, err := nextSeqID(tx, "run_seq", "R")
		if err != nil {
			return err
		}
		snapshot := p.ActorSnapshot
		if snapshot == "" {
			snapshot = "{}"
		}
		r = &Run{
			ID: runID, TicketID: p.TicketID, ActorID: p.ActorID, ActorSnapshot: snapshot,
			Mode: string(p.Mode), Runtime: p.Runtime, Model: p.Model,
			Status: string(run.StatusRunning), WorkspacePath: p.WorkspacePath,
			BaseCommit: p.BaseCommit, StartedAt: nowStr, UpdatedAt: nowStr,
		}
		if _, err := tx.Exec(`INSERT INTO runs
			(id, ticket_id, actor_id, actor_snapshot_json, mode, runtime, model, status,
			 workspace_path, base_commit, started_at, updated_at)
			VALUES (?,?,?,?,?,?,?,?,?,?,?,?)`,
			r.ID, r.TicketID, r.ActorID, r.ActorSnapshot, r.Mode, r.Runtime, nullStr(r.Model),
			r.Status, r.WorkspacePath, nullStr(r.BaseCommit), r.StartedAt, r.UpdatedAt); err != nil {
			return err
		}
		// Claim the node with this run id (single-winner) in the same transaction.
		l, err := txClaim(tx, p.TicketID, runID, p.ActorID, now, p.TTL)
		if err != nil {
			return err
		}
		lease = l
		return appendAudit(tx, p.ActorID, "run.started", "run", runID, map[string]any{
			"ticket_id": p.TicketID, "mode": r.Mode,
		})
	})
	if err != nil {
		return nil, nil, err
	}
	return r, lease, nil
}

// SetRunStatus transitions a run to a new status, validating the move against
// the run lifecycle (ADR 0027) and appending an audit event. Reaching a terminal
// status stamps completed_at.
func (db *DB) SetRunStatus(runID string, to run.Status, actor string) error {
	if !to.Valid() {
		return fmt.Errorf("invalid run status %q", to)
	}
	now := encoding.FormatTime(time.Now())
	return db.withTx(func(tx *sql.Tx) error {
		var fromStr string
		err := tx.QueryRow(`SELECT status FROM runs WHERE id=?`, runID).Scan(&fromStr)
		if err == sql.ErrNoRows {
			return ErrNotFound
		}
		if err != nil {
			return err
		}
		from := run.Status(fromStr)
		if !run.CanTransition(from, to) {
			return fmt.Errorf("%w: run %s -> %s", ErrIllegalTransition, from, to)
		}
		if to.Terminal() {
			_, err = tx.Exec(`UPDATE runs SET status=?, updated_at=?, completed_at=? WHERE id=?`,
				string(to), now, now, runID)
		} else {
			_, err = tx.Exec(`UPDATE runs SET status=?, updated_at=? WHERE id=?`, string(to), now, runID)
		}
		if err != nil {
			return err
		}
		return appendAudit(tx, actor, "run.status", "run", runID, map[string]any{
			"from": from, "to": to,
		})
	})
}

// CancelRun cancels a run: it transitions the run to cancelled, releases the
// node's lease, and (if the node is still in_progress) moves it to blocked so a
// human can requeue it. All in one transaction with audit events.
func (db *DB) CancelRun(runID, actor string) error {
	now := encoding.FormatTime(time.Now())
	return db.withTx(func(tx *sql.Tx) error {
		var ticketID, statusStr string
		err := tx.QueryRow(`SELECT ticket_id, status FROM runs WHERE id=?`, runID).Scan(&ticketID, &statusStr)
		if err == sql.ErrNoRows {
			return ErrNotFound
		}
		if err != nil {
			return err
		}
		from := run.Status(statusStr)
		if !run.CanTransition(from, run.StatusCancelled) {
			return fmt.Errorf("%w: run %s -> cancelled", ErrIllegalTransition, from)
		}
		if _, err := tx.Exec(`UPDATE runs SET status=?, updated_at=?, completed_at=? WHERE id=?`,
			string(run.StatusCancelled), now, now, runID); err != nil {
			return err
		}
		if _, err := tx.Exec(`DELETE FROM leases WHERE ticket_id=? AND run_id=?`, ticketID, runID); err != nil {
			return err
		}
		// Return the node to blocked if it was being worked, so it is not left
		// stranded in_progress with no active run.
		var tStatus string
		if err := tx.QueryRow(`SELECT status FROM tickets WHERE id=?`, ticketID).Scan(&tStatus); err == nil {
			if ticket.Status(tStatus) == ticket.StatusInProgress {
				if _, err := tx.Exec(`UPDATE tickets SET status=?, updated_at=? WHERE id=?`,
					string(ticket.StatusBlocked), now, ticketID); err != nil {
					return err
				}
				if err := appendAudit(tx, actor, "ticket.transitioned", "ticket", ticketID, map[string]any{
					"from": ticket.StatusInProgress, "to": ticket.StatusBlocked, "reason": "run_cancelled",
				}); err != nil {
					return err
				}
			}
		}
		return appendAudit(tx, actor, "run.status", "run", runID, map[string]any{"from": from, "to": run.StatusCancelled})
	})
}

// GetRun returns one run, or ErrNotFound.
func (db *DB) GetRun(id string) (*Run, error) {
	return scanRun(db.QueryRow(`SELECT `+runColumns+` FROM runs WHERE id = ?`, id), id)
}

// ListRuns returns runs, newest first.
func (db *DB) ListRuns() ([]*Run, error) {
	rows, err := db.Query(`SELECT ` + runColumns + ` FROM runs ORDER BY id DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanRuns(rows)
}

// ListRunsForTicket returns a node's runs, newest first.
func (db *DB) ListRunsForTicket(ticketID string) ([]*Run, error) {
	rows, err := db.Query(`SELECT `+runColumns+` FROM runs WHERE ticket_id=? ORDER BY id DESC`, ticketID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanRuns(rows)
}

func scanRuns(rows *sql.Rows) ([]*Run, error) {
	out := []*Run{}
	for rows.Next() {
		r, err := scanRun(rows, "")
		if err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

func scanRun(s rowScanner, id string) (*Run, error) {
	var (
		r           Run
		model       sql.NullString
		baseCommit  sql.NullString
		completedAt sql.NullString
		lastEvent   sql.NullString
		lastMessage sql.NullString
	)
	err := s.Scan(&r.ID, &r.TicketID, &r.ActorID, &r.ActorSnapshot, &r.Mode, &r.Runtime, &model,
		&r.Status, &r.WorkspacePath, &baseCommit, &r.StartedAt, &r.UpdatedAt, &completedAt,
		&lastEvent, &lastMessage, &r.InputTokens, &r.OutputTokens, &r.TotalTokens)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	r.Model = model.String
	r.BaseCommit = baseCommit.String
	r.CompletedAt = completedAt.String
	r.LastEvent = lastEvent.String
	r.LastMessage = lastMessage.String
	return &r, nil
}
