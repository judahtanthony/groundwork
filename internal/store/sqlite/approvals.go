package sqlite

import (
	"database/sql"
	"fmt"
	"strconv"

	"groundwork/internal/approval"
	"groundwork/internal/encoding"
)

// Approval is a recorded capability gate (docs/contracts/sqlite-schema.md).
type Approval struct {
	ID               string   `json:"id"`
	RunID            string   `json:"run_id,omitempty"`
	TicketID         string   `json:"ticket_id"`
	Type             string   `json:"type"`
	RiskClass        string   `json:"risk_class"`
	RiskScore        *int     `json:"risk_score,omitempty"`
	Reversible       *bool    `json:"reversible,omitempty"`
	Summary          string   `json:"summary"`
	ActionJSON       string   `json:"action_json"`
	Status           string   `json:"status"`
	RequestedByActor string   `json:"requested_by_actor"`
	DecidedByActor   string   `json:"decided_by_actor,omitempty"`
	RequiredActors   []string `json:"required_actors"`
	RequiredRoles    []string `json:"required_roles"`
	DecisionReason   string   `json:"decision_reason,omitempty"`
	CreatedAt        string   `json:"created_at"`
	DecidedAt        string   `json:"decided_at,omitempty"`
}

// CreateApprovalParams configures a new approval record.
type CreateApprovalParams struct {
	RunID            string
	TicketID         string
	Type             approval.Type
	RiskClass        string
	RiskScore        *int
	Reversible       *bool
	Summary          string
	ActionJSON       string
	Status           approval.Status // pending, or approved when policy auto-approves
	RequestedByActor string
	DecidedByActor   string // set when auto-approved by policy
	DecisionReason   string
	RequiredActors   []string
	RequiredRoles    []string
}

const approvalColumns = `id, run_id, ticket_id, type, risk_class, risk_score, reversible,
	summary, action_json, status, requested_by_actor, decided_by_actor,
	required_actors_json, required_roles_json, decision_reason, created_at, decided_at`

// CreateApproval inserts an approval record (pending, or already-approved when
// policy auto-approved it) and appends an audit event.
func (db *DB) CreateApproval(p CreateApprovalParams) (*Approval, error) {
	var a *Approval
	err := db.withTx(func(tx *sql.Tx) error {
		created, err := txCreateApproval(tx, p)
		a = created
		return err
	})
	if err != nil {
		return nil, err
	}
	return a, nil
}

// txCreateApproval inserts an approval within tx and appends its audit event.
// It is the single approval-insert path shared by CreateApproval,
// DecomposeProposal, and Escalate, so the column list lives in one place.
func txCreateApproval(tx *sql.Tx, p CreateApprovalParams) (*Approval, error) {
	if !p.Type.Valid() {
		return nil, fmt.Errorf("invalid approval type %q", p.Type)
	}
	if p.Status == "" {
		p.Status = approval.StatusPending
	}
	now := encoding.Now()
	actionJSON := p.ActionJSON
	if actionJSON == "" {
		actionJSON = "{}"
	}
	reqActors, _ := encoding.JSON(orEmpty(p.RequiredActors))
	reqRoles, _ := encoding.JSON(orEmpty(p.RequiredRoles))

	id, err := nextSeqID(tx, "approval_seq", "A")
	if err != nil {
		return nil, err
	}
	decidedAt := ""
	if approval.Status(p.Status).Terminal() {
		decidedAt = now
	}
	if _, err := tx.Exec(`INSERT INTO approvals
		(id, run_id, ticket_id, type, risk_class, risk_score, reversible, summary, action_json,
		 status, requested_by_actor, decided_by_actor, required_actors_json, required_roles_json,
		 decision_reason, created_at, decided_at)
		VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		id, nullStr(p.RunID), p.TicketID, string(p.Type), p.RiskClass, nullIntPtr(p.RiskScore),
		nullBool(p.Reversible), p.Summary, actionJSON, string(p.Status), p.RequestedByActor,
		nullStr(p.DecidedByActor), reqActors, reqRoles, nullStr(p.DecisionReason), now, nullStr(decidedAt)); err != nil {
		return nil, err
	}
	if err := appendAudit(tx, p.RequestedByActor, "approval.created", "approval", id, map[string]any{
		"ticket_id": p.TicketID, "type": p.Type, "status": p.Status,
	}); err != nil {
		return nil, err
	}
	return &Approval{
		ID: id, RunID: p.RunID, TicketID: p.TicketID, Type: string(p.Type), RiskClass: p.RiskClass,
		RiskScore: p.RiskScore, Reversible: p.Reversible, Summary: p.Summary, ActionJSON: actionJSON,
		Status: string(p.Status), RequestedByActor: p.RequestedByActor, DecidedByActor: p.DecidedByActor,
		RequiredActors: orEmpty(p.RequiredActors), RequiredRoles: orEmpty(p.RequiredRoles),
		DecisionReason: p.DecisionReason, CreatedAt: now, DecidedAt: decidedAt,
	}, nil
}

// DecideApproval records a terminal or clarifying decision on a pending approval,
// validating the transition and appending an audit event. It is the generic path
// (execute, land_to_main); decompose/replan carry side effects in dedicated
// methods.
func (db *DB) DecideApproval(id string, to approval.Status, decidedBy, reason string) (*Approval, error) {
	if !to.Valid() {
		return nil, fmt.Errorf("invalid approval status %q", to)
	}
	var a *Approval
	err := db.withTx(func(tx *sql.Tx) error {
		cur, err := txGetApproval(tx, id)
		if err != nil {
			return err
		}
		if approval.Status(cur.Status) != approval.StatusPending {
			return fmt.Errorf("%w: approval already %s", ErrIllegalTransition, cur.Status)
		}
		if err := txSetApprovalDecision(tx, id, to, decidedBy, reason); err != nil {
			return err
		}
		cur.Status = string(to)
		cur.DecidedByActor = decidedBy
		cur.DecisionReason = reason
		a = cur
		return nil
	})
	if err != nil {
		return nil, err
	}
	return a, nil
}

// txSetApprovalDecision updates an approval's decision fields and audits it.
func txSetApprovalDecision(tx *sql.Tx, id string, to approval.Status, decidedBy, reason string) error {
	now := encoding.Now()
	decidedAt := ""
	if to.Terminal() {
		decidedAt = now
	}
	if _, err := tx.Exec(`UPDATE approvals SET status=?, decided_by_actor=?, decision_reason=?, decided_at=? WHERE id=?`,
		string(to), nullStr(decidedBy), nullStr(reason), nullStr(decidedAt), id); err != nil {
		return err
	}
	return appendAudit(tx, decidedBy, "approval.decided", "approval", id, map[string]any{
		"status": to, "reason": reason,
	})
}

// GetApproval returns one approval, or ErrNotFound.
func (db *DB) GetApproval(id string) (*Approval, error) {
	return scanApproval(db.QueryRow(`SELECT `+approvalColumns+` FROM approvals WHERE id=?`, id))
}

// ListApprovals returns approvals newest-first, optionally filtered by status.
func (db *DB) ListApprovals(status string) ([]*Approval, error) {
	q := `SELECT ` + approvalColumns + ` FROM approvals`
	args := []any{}
	if status != "" {
		q += ` WHERE status=?`
		args = append(args, status)
	}
	q += ` ORDER BY id DESC`
	rows, err := db.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []*Approval{}
	for rows.Next() {
		a, err := scanApproval(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

// HasOpenApprovalOfType reports whether a node already has a pending approval of
// the given type. Used to dedup repeatedly-raised exceptions so a node that keeps
// failing a gate does not flood the queue (one open request per node/type).
func (db *DB) HasOpenApprovalOfType(ticketID string, typ approval.Type) (bool, error) {
	var n int
	err := db.QueryRow(`SELECT COUNT(*) FROM approvals WHERE ticket_id=? AND type=? AND status=?`,
		ticketID, string(typ), string(approval.StatusPending)).Scan(&n)
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

// txGetApproval loads an approval within a transaction.
func txGetApproval(tx *sql.Tx, id string) (*Approval, error) {
	return scanApproval(tx.QueryRow(`SELECT `+approvalColumns+` FROM approvals WHERE id=?`, id))
}

func scanApproval(s rowScanner) (*Approval, error) {
	var (
		a              Approval
		runID          sql.NullString
		riskScore      sql.NullInt64
		reversible     sql.NullBool
		decidedBy      sql.NullString
		decisionReason sql.NullString
		decidedAt      sql.NullString
		reqActors      string
		reqRoles       string
	)
	err := s.Scan(&a.ID, &runID, &a.TicketID, &a.Type, &a.RiskClass, &riskScore, &reversible,
		&a.Summary, &a.ActionJSON, &a.Status, &a.RequestedByActor, &decidedBy,
		&reqActors, &reqRoles, &decisionReason, &a.CreatedAt, &decidedAt)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	a.RunID = runID.String
	a.DecidedByActor = decidedBy.String
	a.DecisionReason = decisionReason.String
	a.DecidedAt = decidedAt.String
	if riskScore.Valid {
		v := int(riskScore.Int64)
		a.RiskScore = &v
	}
	if reversible.Valid {
		a.Reversible = &reversible.Bool
	}
	if err := unmarshalList(reqActors, &a.RequiredActors); err != nil {
		return nil, err
	}
	if err := unmarshalList(reqRoles, &a.RequiredRoles); err != nil {
		return nil, err
	}
	return &a, nil
}

// nextSeqID allocates the next monotonic id for a meta sequence key, formatted
// as "<prefix>-0001".
func nextSeqID(tx *sql.Tx, key, prefix string) (string, error) {
	var cur int
	var s string
	err := tx.QueryRow(`SELECT value FROM meta WHERE key=?`, key).Scan(&s)
	switch {
	case err == sql.ErrNoRows:
		cur = 0
	case err != nil:
		return "", err
	default:
		if cur, err = strconv.Atoi(s); err != nil {
			return "", fmt.Errorf("corrupt %s %q: %w", key, s, err)
		}
	}
	next := cur + 1
	if _, err := tx.Exec(
		`INSERT INTO meta (key, value) VALUES (?, ?) ON CONFLICT(key) DO UPDATE SET value = excluded.value`,
		key, strconv.Itoa(next),
	); err != nil {
		return "", err
	}
	return fmt.Sprintf("%s-%04d", prefix, next), nil
}

func orEmpty(s []string) []string {
	if s == nil {
		return []string{}
	}
	return s
}

func nullBool(b *bool) any {
	if b == nil {
		return nil
	}
	return *b
}

func nullIntPtr(p *int) any {
	if p == nil {
		return nil
	}
	return *p
}
