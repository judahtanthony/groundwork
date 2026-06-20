package sqlite

import (
	"database/sql"

	"groundwork/internal/approval"
	"groundwork/internal/encoding"
	"groundwork/internal/ticket"
)

// Escalate records a typed upward-revision escalation on a node: it moves the
// node to blocked and opens a pending re-plan approval routed for a human
// decision (work-tree.md, ADR 0010/0024). The re-plan is human-gated in v1; the
// automatic sibling-rework cascade is deferred.
func (db *DB) Escalate(ticketID, reason, actor string) (*Approval, error) {
	if reason == "" {
		reason = "escalation"
	}
	now := encoding.Now()
	var appr *Approval
	err := db.withTx(func(tx *sql.Tx) error {
		var status string
		err := tx.QueryRow(`SELECT status FROM tickets WHERE id=?`, ticketID).Scan(&status)
		if err == sql.ErrNoRows {
			return ErrNotFound
		}
		if err != nil {
			return err
		}
		// Move the node to blocked (coordinator-driven state set).
		if _, err := tx.Exec(`UPDATE tickets SET status=?, updated_at=? WHERE id=?`,
			string(ticket.StatusBlocked), now, ticketID); err != nil {
			return err
		}
		if err := appendAudit(tx, actor, "ticket.escalated", "ticket", ticketID, map[string]any{
			"reason": reason, "from": status,
		}); err != nil {
			return err
		}

		// Link the most recent run, if any, for traceability.
		var runID sql.NullString
		_ = tx.QueryRow(`SELECT id FROM runs WHERE ticket_id=? ORDER BY id DESC LIMIT 1`, ticketID).Scan(&runID)

		actionJSON, _ := encoding.JSON(map[string]any{"reason": reason})
		appr, err = txCreateApproval(tx, CreateApprovalParams{
			RunID: runID.String, TicketID: ticketID, Type: approval.TypeReplan, RiskClass: string(riskLow),
			Summary: "Re-plan: " + reason, ActionJSON: actionJSON, RequestedByActor: actor,
		})
		return err
	})
	if err != nil {
		return nil, err
	}
	return appr, nil
}

// AcceptReplan approves a re-plan: it records the decision and requeues the
// escalated node (blocked -> todo) so the scheduler can dispatch it against the
// revised plan. This is the seam where a node stalled behind a cancelled
// prerequisite (ADR 0024) is resolved.
func (db *DB) AcceptReplan(approvalID, decidedBy, reason string) (*Approval, error) {
	return db.replanDecision(approvalID, decidedBy, reason, true)
}

// RejectReplan rejects a re-plan: the node stays blocked.
func (db *DB) RejectReplan(approvalID, decidedBy, reason string) (*Approval, error) {
	return db.replanDecision(approvalID, decidedBy, reason, false)
}

func (db *DB) replanDecision(approvalID, decidedBy, reason string, accept bool) (*Approval, error) {
	var appr *Approval
	err := db.withTx(func(tx *sql.Tx) error {
		a, err := txGetApproval(tx, approvalID)
		if err != nil {
			return err
		}
		if err := requireDecidable(a, approval.TypeReplan); err != nil {
			return err
		}
		if accept {
			now := encoding.Now()
			var status string
			if err := tx.QueryRow(`SELECT status FROM tickets WHERE id=?`, a.TicketID).Scan(&status); err != nil {
				return err
			}
			if ticket.Status(status) == ticket.StatusBlocked {
				if _, err := tx.Exec(`UPDATE tickets SET status=?, updated_at=? WHERE id=?`,
					string(ticket.StatusTodo), now, a.TicketID); err != nil {
					return err
				}
				if err := appendAudit(tx, decidedBy, "ticket.transitioned", "ticket", a.TicketID, map[string]any{
					"to": ticket.StatusTodo, "reason": "replan_accepted",
				}); err != nil {
					return err
				}
			}
			if err := txSetApprovalDecision(tx, approvalID, approval.StatusApproved, decidedBy, reason); err != nil {
				return err
			}
		} else {
			if err := txSetApprovalDecision(tx, approvalID, approval.StatusRejected, decidedBy, reason); err != nil {
				return err
			}
		}
		appr, err = txGetApproval(tx, approvalID)
		return err
	})
	if err != nil {
		return nil, err
	}
	return appr, nil
}
