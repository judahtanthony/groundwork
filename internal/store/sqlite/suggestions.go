package sqlite

import (
	"database/sql"
	"fmt"

	"groundwork/internal/encoding"
	"groundwork/internal/ticket"
)

// PolicySuggestion is an advisory, human-reviewed proposal to elevate autonomy
// (ADR 0038). The system creates these; a human promotes or dismisses them. The
// system never self-elevates: promotion records the decision and emits the policy
// change for a human to apply (amend_policy/elevate_autonomy stay human-gated).
type PolicySuggestion struct {
	ID         string `json:"id"`
	Kind       string `json:"kind"`
	ActionType string `json:"action_type"`
	WorkType   string `json:"work_type"`
	Level      string `json:"level"`
	Rationale  string `json:"rationale"`
	Status     string `json:"status"`
	CreatedAt  string `json:"created_at"`
	DecidedAt  string `json:"decided_at,omitempty"`
}

// SuggestionElevationThreshold is how many clean done leaves of a work type are
// needed before the system proposes elevating that work type's execute autonomy.
const SuggestionElevationThreshold = 3

const suggestionColumns = `id, kind, action_type, work_type, level, rationale, status, created_at, decided_at`

func scanSuggestion(s rowScanner) (*PolicySuggestion, error) {
	var p PolicySuggestion
	var decided sql.NullString
	if err := s.Scan(&p.ID, &p.Kind, &p.ActionType, &p.WorkType, &p.Level, &p.Rationale, &p.Status, &p.CreatedAt, &decided); err != nil {
		return nil, err
	}
	p.DecidedAt = decided.String
	return &p, nil
}

// ListSuggestions returns suggestions newest-first, optionally filtered by status.
func (db *DB) ListSuggestions(status string) ([]*PolicySuggestion, error) {
	q := `SELECT ` + suggestionColumns + ` FROM policy_suggestions`
	var args []any
	if status != "" {
		q += ` WHERE status = ?`
		args = append(args, status)
	}
	q += ` ORDER BY id DESC`
	rows, err := db.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []*PolicySuggestion{}
	for rows.Next() {
		s, err := scanSuggestion(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

// GetSuggestion returns one suggestion or ErrNotFound.
func (db *DB) GetSuggestion(id string) (*PolicySuggestion, error) {
	row := db.QueryRow(`SELECT `+suggestionColumns+` FROM policy_suggestions WHERE id = ?`, id)
	s, err := scanSuggestion(row)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	return s, err
}

// SetSuggestionStatus records a human decision (promoted | dismissed). It never
// changes policy itself — applying a promoted elevation is a separate human-gated
// amend_policy action.
func (db *DB) SetSuggestionStatus(id, status string) (*PolicySuggestion, error) {
	res, err := db.Exec(`UPDATE policy_suggestions SET status = ?, decided_at = ? WHERE id = ?`,
		status, encoding.Now(), id)
	if err != nil {
		return nil, err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return nil, ErrNotFound
	}
	return db.GetSuggestion(id)
}

// GenerateElevationSuggestions scans the work tree and proposes elevating execute
// autonomy to "auto" for any work type with a clean track record — at least
// SuggestionElevationThreshold done leaves, each with a passing validation and no
// failures — that has no pending suggestion yet. It only proposes; it never
// elevates. Returns the suggestions created this scan.
func (db *DB) GenerateElevationSuggestions() ([]*PolicySuggestion, error) {
	all, err := db.ListTickets()
	if err != nil {
		return nil, err
	}
	type counts struct {
		clean   int
		anyFail bool
	}
	byWT := map[string]*counts{}
	for _, t := range all {
		if t.NodeType != ticket.NodeLeaf || t.Status != ticket.StatusDone || t.WorkType == "" {
			continue
		}
		vs, err := db.ListValidationsForTicket(t.ID)
		if err != nil {
			return nil, err
		}
		hasPass, hasFail := false, false
		for _, v := range vs {
			switch v.Status {
			case ValidationPass:
				hasPass = true
			case ValidationFail:
				hasFail = true
			}
		}
		c := byWT[t.WorkType]
		if c == nil {
			c = &counts{}
			byWT[t.WorkType] = c
		}
		if hasFail {
			c.anyFail = true
		} else if hasPass {
			c.clean++
		}
	}

	var created []*PolicySuggestion
	for wt, c := range byWT {
		if c.anyFail || c.clean < SuggestionElevationThreshold {
			continue
		}
		rationale := fmt.Sprintf("%d done %s leaves landed with passing validation and no failures",
			c.clean, wt)
		s, err := db.createElevationSuggestionIfAbsent(wt, rationale)
		if err != nil {
			return nil, err
		}
		if s != nil {
			created = append(created, s)
		}
	}
	return created, nil
}

// createElevationSuggestionIfAbsent inserts a pending execute-elevation suggestion
// for workType, or returns (nil, nil) if one is already pending (the partial
// unique index also enforces this).
func (db *DB) createElevationSuggestionIfAbsent(workType, rationale string) (*PolicySuggestion, error) {
	var s *PolicySuggestion
	err := db.withTx(func(tx *sql.Tx) error {
		var existing int
		if err := tx.QueryRow(`SELECT COUNT(*) FROM policy_suggestions
			WHERE kind = 'elevate_autonomy' AND action_type = 'execute' AND work_type = ? AND status = 'pending'`,
			workType).Scan(&existing); err != nil {
			return err
		}
		if existing > 0 {
			return nil
		}
		id, err := nextSeqID(tx, "suggestion_seq", "S")
		if err != nil {
			return err
		}
		now := encoding.Now()
		if _, err := tx.Exec(`INSERT INTO policy_suggestions
			(id, kind, action_type, work_type, level, rationale, status, created_at)
			VALUES (?, 'elevate_autonomy', 'execute', ?, 'auto', ?, 'pending', ?)`,
			id, workType, rationale, now); err != nil {
			return err
		}
		s = &PolicySuggestion{
			ID: id, Kind: "elevate_autonomy", ActionType: "execute", WorkType: workType,
			Level: "auto", Rationale: rationale, Status: "pending", CreatedAt: now,
		}
		return nil
	})
	return s, err
}
