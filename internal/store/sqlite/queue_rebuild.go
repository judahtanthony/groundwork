package sqlite

import (
	"database/sql"

	"groundwork/internal/approval"
	"groundwork/internal/decision"
	"groundwork/internal/encoding"
	"groundwork/internal/ticket"
)

// QueueRebuildReport summarizes what RebuildDurableQueues projected.
type QueueRebuildReport struct {
	ApprovalsRecreated int `json:"approvals_recreated"`
	RecoveryNeeded     int `json:"recovery_needed"`
}

// RebuildDurableQueues projects pending durable decision records (ADR 0051) into
// the live coordinator queues after a cold rebuild, and surfaces recovery_needed
// for stranded tickets. It runs at startup after cold import and ReconcileStartup.
//
// Two passes:
//
//  1. Each pending approval_requested record without a matching live approval row
//     recreates one with a fresh runtime handle (the durable id is stable; the
//     A-id is a disposable live handle, ADR 0051). input_requested and
//     decision_requested records need no table — they are themselves the durable
//     explainer for a blocked ticket and project as a query over pending records.
//
//  2. Any blocked/review/rework ticket with no durable explainer and no live
//     pending approval gets a recovery_needed record appended, rather than being
//     left silently stranded (ADR 0051, decision-records.md).
//
// The pass is idempotent: a recreated approval and an appended recovery_needed
// record both become explainers, so a second run is a no-op.
func (db *DB) RebuildDurableQueues() (*QueueRebuildReport, error) {
	rep := &QueueRebuildReport{}

	pending, err := db.ListPendingDecisions()
	if err != nil {
		return nil, err
	}

	// Pass 1: recreate approval rows from pending approval_requested records.
	liveApprovals, err := db.pendingApprovalKeys()
	if err != nil {
		return nil, err
	}
	for _, rec := range pending {
		if rec.EventType != decision.EventApprovalRequested {
			continue
		}
		typ := approval.Type(rec.RequestType)
		if !typ.Valid() {
			continue // unknown request type: cannot recreate a typed approval row
		}
		key := rec.TicketID + "\x00" + string(typ)
		if liveApprovals[key] {
			continue // already projected
		}
		if _, err := db.CreateApproval(approvalFromRecord(rec, typ)); err != nil {
			return nil, err
		}
		liveApprovals[key] = true
		rep.ApprovalsRecreated++
	}

	// Pass 2: surface recovery_needed for stranded tickets. Recompute the
	// explainer sets after pass 1 so recreated approvals count.
	explained, err := db.ticketsWithPendingDecision()
	if err != nil {
		return nil, err
	}
	approved, err := db.ticketsWithPendingApproval()
	if err != nil {
		return nil, err
	}
	stranded, err := db.ticketsInStatuses(ticket.StatusBlocked, ticket.StatusReview, ticket.StatusRework)
	if err != nil {
		return nil, err
	}
	for _, id := range stranded {
		if explained[id] || approved[id] {
			continue
		}
		if _, err := db.AppendDecision(decision.Record{
			EventType:      decision.EventRecoveryNeeded,
			TicketID:       id,
			Status:         decision.StatusPending,
			RequestedAt:    encoding.Now(),
			HandoffSummary: "recovery_needed: ticket status has no durable blocker or pending request after rebuild",
		}); err != nil {
			return nil, err
		}
		rep.RecoveryNeeded++
	}
	return rep, nil
}

// approvalFromRecord maps a durable approval_requested record to the params for a
// recreated live approval row.
func approvalFromRecord(rec decision.Record, typ approval.Type) CreateApprovalParams {
	summary := rec.Statement
	if summary == "" {
		summary = rec.HandoffSummary
	}
	p := CreateApprovalParams{
		TicketID:         rec.TicketID,
		RunID:            rec.RunID,
		Type:             typ,
		Summary:          summary,
		Status:           approval.StatusPending,
		RequestedByActor: rec.RequestedBy,
		RequiredRoles:    rec.RequiredRoles,
	}
	if rec.RequestedActor != "" {
		p.RequiredActors = []string{rec.RequestedActor}
	}
	if pi := rec.PolicyInputs; pi != nil {
		p.RiskClass = pi.RiskClass
		p.Reversible = pi.Reversible
	}
	return p
}

// pendingApprovalKeys returns the set of "ticket\x00type" keys with a live
// pending approval, so recreation skips already-projected requests.
func (db *DB) pendingApprovalKeys() (map[string]bool, error) {
	rows, err := db.Query(`SELECT ticket_id, type FROM approvals WHERE status = ?`, string(approval.StatusPending))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[string]bool{}
	for rows.Next() {
		var t, typ string
		if err := rows.Scan(&t, &typ); err != nil {
			return nil, err
		}
		out[t+"\x00"+typ] = true
	}
	return out, rows.Err()
}

// ticketsWithPendingApproval returns the set of ticket ids that have any pending
// approval row.
func (db *DB) ticketsWithPendingApproval() (map[string]bool, error) {
	rows, err := db.Query(`SELECT DISTINCT ticket_id FROM approvals WHERE status = ?`, string(approval.StatusPending))
	if err != nil {
		return nil, err
	}
	return scanIDSet(rows)
}

// ticketsWithPendingDecision returns the set of ticket ids that have any pending
// durable decision record (the durable explainer for a blocked/review ticket).
func (db *DB) ticketsWithPendingDecision() (map[string]bool, error) {
	rows, err := db.Query(`SELECT DISTINCT ticket_id FROM decisions WHERE status = ?`, decision.StatusPending)
	if err != nil {
		return nil, err
	}
	return scanIDSet(rows)
}

// ticketsInStatuses returns ticket ids currently in any of the given statuses.
func (db *DB) ticketsInStatuses(statuses ...ticket.Status) ([]string, error) {
	if len(statuses) == 0 {
		return nil, nil
	}
	query := `SELECT id FROM tickets WHERE status IN (`
	args := make([]any, len(statuses))
	for i, s := range statuses {
		if i > 0 {
			query += ","
		}
		query += "?"
		args[i] = string(s)
	}
	query += `) ORDER BY id`
	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

func scanIDSet(rows *sql.Rows) (map[string]bool, error) {
	defer rows.Close()
	out := map[string]bool{}
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		out[id] = true
	}
	return out, rows.Err()
}
