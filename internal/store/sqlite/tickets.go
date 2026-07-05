package sqlite

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"groundwork/internal/encoding"
	"groundwork/internal/ticket"
)

// ErrNotFound is returned when a requested row does not exist.
var ErrNotFound = errors.New("not found")

// ticketColumns is the canonical column order for ticket SELECTs.
const ticketColumns = `id, parent_id, kind, node_type, work_type, title, description,
	contract_json, status, assignee, requested_actor, priority, labels_json,
	acceptance_json, risk_score, created_at, updated_at`

// withTx runs fn inside a transaction, committing on success.
func (db *DB) withTx(fn func(*sql.Tx) error) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if err := fn(tx); err != nil {
		return err
	}
	return tx.Commit()
}

// nextTicketID allocates the next monotonic ticket id (ADR 0019) within tx.
func nextTicketID(tx *sql.Tx) (string, error) {
	var cur int
	var s string
	err := tx.QueryRow(`SELECT value FROM meta WHERE key='ticket_seq'`).Scan(&s)
	switch {
	case err == sql.ErrNoRows:
		cur = 0
	case err != nil:
		return "", err
	default:
		if cur, err = strconv.Atoi(s); err != nil {
			return "", fmt.Errorf("corrupt ticket_seq %q: %w", s, err)
		}
	}
	next := cur + 1
	if _, err := tx.Exec(
		`INSERT INTO meta (key, value) VALUES ('ticket_seq', ?)
		 ON CONFLICT(key) DO UPDATE SET value = excluded.value`,
		strconv.Itoa(next),
	); err != nil {
		return "", err
	}
	return fmt.Sprintf("T-%04d", next), nil
}

// SeedTicketSeq raises the ticket sequence to at least n, so imported ids are
// not reused (ADR 0019). It never lowers the sequence.
func (db *DB) SeedTicketSeq(n int) error {
	return db.withTx(func(tx *sql.Tx) error {
		var cur int
		var s string
		err := tx.QueryRow(`SELECT value FROM meta WHERE key='ticket_seq'`).Scan(&s)
		switch {
		case err == nil:
			if cur, err = strconv.Atoi(s); err != nil {
				return fmt.Errorf("corrupt ticket_seq %q: %w", s, err)
			}
		case err != sql.ErrNoRows:
			return err
		}
		if n <= cur {
			return nil
		}
		_, err = tx.Exec(
			`INSERT INTO meta (key, value) VALUES ('ticket_seq', ?)
			 ON CONFLICT(key) DO UPDATE SET value = excluded.value`,
			strconv.Itoa(n),
		)
		return err
	})
}

// CreateTicket inserts t, allocating an id and timestamps and applying defaults.
// It appends a ticket.created audit event in the same transaction. The assigned
// id and timestamps are written back into t.
func (db *DB) CreateTicket(t *ticket.Ticket, actor string) error {
	if err := prepareTicket(t); err != nil {
		return err
	}
	now := encoding.Now()
	t.CreatedAt, t.UpdatedAt = now, now

	if err := db.withTx(func(tx *sql.Tx) error {
		if t.ID == "" {
			id, err := nextTicketID(tx)
			if err != nil {
				return err
			}
			t.ID = id
		}

		labels, acceptance, err := marshalLists(t)
		if err != nil {
			return err
		}

		_, err = tx.Exec(`INSERT INTO tickets
			(id, parent_id, kind, node_type, work_type, title, description, contract_json,
			 status, assignee, requested_actor, priority, labels_json, acceptance_json,
			 risk_score, created_at, updated_at)
			VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
			t.ID, nullStr(t.ParentID), t.Kind, nullStr(string(t.NodeType)), nullStr(t.WorkType),
			t.Title, t.Description, t.Contract, string(t.Status), nullStr(t.Assignee),
			nullStr(t.RequestedActor), nullFloat(t.Priority), labels, acceptance,
			nullInt(t.RiskScore), t.CreatedAt, t.UpdatedAt)
		if err != nil {
			return err
		}
		return appendAudit(tx, actor, "ticket.created", "ticket", t.ID, map[string]any{
			"title":  t.Title,
			"status": t.Status,
		})
	}); err != nil {
		return err
	}
	return db.writeThrough(t.ID)
}

// GetTicket returns the ticket with the given id, or ErrNotFound.
func (db *DB) GetTicket(id string) (*ticket.Ticket, error) {
	row := db.QueryRow(`SELECT `+ticketColumns+` FROM tickets WHERE id = ?`, id)
	return scanTicket(row)
}

// ListTickets returns all tickets ordered by id.
func (db *DB) ListTickets() ([]*ticket.Ticket, error) {
	rows, err := db.Query(`SELECT ` + ticketColumns + ` FROM tickets ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanTickets(rows)
}

// scanTickets drains rows of ticketColumns into tickets.
func scanTickets(rows *sql.Rows) ([]*ticket.Ticket, error) {
	var out []*ticket.Ticket
	for rows.Next() {
		t, err := scanTicket(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

// UpdateTicket persists mutable fields of t and refreshes updated_at, appending
// a ticket.updated audit event. It deliberately does NOT change status or
// parent_id: status moves only through TransitionTicket (ADR 0022), and parentage
// is fixed at creation in Phase 1 (no reparent command), which also keeps the
// parent tree acyclic by construction.
func (db *DB) UpdateTicket(t *ticket.Ticket, actor string) error {
	if err := prepareTicket(t); err != nil {
		return err
	}
	t.UpdatedAt = encoding.Now()
	if err := db.withTx(func(tx *sql.Tx) error {
		labels, acceptance, err := marshalLists(t)
		if err != nil {
			return err
		}
		res, err := tx.Exec(`UPDATE tickets SET
			kind=?, node_type=?, work_type=?, title=?, description=?, contract_json=?,
			assignee=?, requested_actor=?, priority=?, labels_json=?, acceptance_json=?,
			risk_score=?, updated_at=? WHERE id=?`,
			t.Kind, nullStr(string(t.NodeType)), nullStr(t.WorkType), t.Title,
			t.Description, t.Contract, nullStr(t.Assignee), nullStr(t.RequestedActor),
			nullFloat(t.Priority), labels, acceptance, nullInt(t.RiskScore),
			t.UpdatedAt, t.ID)
		if err != nil {
			return err
		}
		if n, _ := res.RowsAffected(); n == 0 {
			return ErrNotFound
		}
		return appendAudit(tx, actor, "ticket.updated", "ticket", t.ID, nil)
	}); err != nil {
		return err
	}
	return db.writeThrough(t.ID)
}

// --- helpers ---

// ErrEmptyTitle is returned when a ticket would be persisted without a title.
var ErrEmptyTitle = errors.New("ticket title must not be empty")

// prepareTicket applies defaults, validates required fields, and canonicalizes
// the contract JSON (ADR 0020) before a write.
func prepareTicket(t *ticket.Ticket) error {
	applyDefaults(t)
	if strings.TrimSpace(t.Title) == "" {
		return ErrEmptyTitle
	}
	canonical, err := canonicalizeJSON(t.Contract)
	if err != nil {
		return fmt.Errorf("invalid contract JSON: %w", err)
	}
	t.Contract = canonical
	return nil
}

// canonicalizeJSON round-trips a JSON document into canonical form (sorted keys,
// no insignificant whitespace) per ADR 0020.
func canonicalizeJSON(s string) (string, error) {
	var v any
	if err := json.Unmarshal([]byte(s), &v); err != nil {
		return "", err
	}
	return encoding.JSON(v)
}

func applyDefaults(t *ticket.Ticket) {
	if t.Kind == "" {
		t.Kind = "ticket"
	}
	if t.Status == "" {
		t.Status = ticket.StatusBacklog
	}
	if t.Contract == "" {
		t.Contract = "{}"
	}
	if t.Labels == nil {
		t.Labels = []string{}
	}
	if t.Acceptance == nil {
		t.Acceptance = []string{}
	}
}

func marshalLists(t *ticket.Ticket) (labels, acceptance string, err error) {
	if labels, err = encoding.JSON(t.Labels); err != nil {
		return
	}
	acceptance, err = encoding.JSON(t.Acceptance)
	return
}

// rowScanner is satisfied by both *sql.Row and *sql.Rows.
type rowScanner interface {
	Scan(dest ...any) error
}

func scanTicket(s rowScanner) (*ticket.Ticket, error) {
	var (
		t              ticket.Ticket
		parentID       sql.NullString
		nodeType       sql.NullString
		workType       sql.NullString
		assignee       sql.NullString
		requestedActor sql.NullString
		priority       sql.NullFloat64
		riskScore      sql.NullInt64
		labels         string
		acceptance     string
	)
	err := s.Scan(&t.ID, &parentID, &t.Kind, &nodeType, &workType, &t.Title, &t.Description,
		&t.Contract, &t.Status, &assignee, &requestedActor, &priority, &labels, &acceptance,
		&riskScore, &t.CreatedAt, &t.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	t.ParentID = parentID.String
	t.NodeType = ticket.NodeType(nodeType.String)
	t.WorkType = workType.String
	t.Assignee = assignee.String
	t.RequestedActor = requestedActor.String
	if priority.Valid {
		p := priority.Float64
		t.Priority = &p
	}
	if riskScore.Valid {
		r := int(riskScore.Int64)
		t.RiskScore = &r
	}
	if err := unmarshalList(labels, &t.Labels); err != nil {
		return nil, err
	}
	if err := unmarshalList(acceptance, &t.Acceptance); err != nil {
		return nil, err
	}
	return &t, nil
}

func unmarshalList(raw string, dst *[]string) error {
	if raw == "" || raw == "null" {
		*dst = []string{}
		return nil
	}
	if err := json.Unmarshal([]byte(raw), dst); err != nil {
		return err
	}
	if *dst == nil {
		*dst = []string{}
	}
	return nil
}

func nullStr(s string) any {
	if s == "" {
		return nil
	}
	return s
}

func nullInt(p *int) any {
	if p == nil {
		return nil
	}
	return *p
}

func nullFloat(p *float64) any {
	if p == nil {
		return nil
	}
	return *p
}
