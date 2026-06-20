package sqlite

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"groundwork/internal/approval"
	"groundwork/internal/encoding"
	"groundwork/internal/run"
	"groundwork/internal/ticket"
)

// ChildSpec is a proposed child node in a decomposition proposal.
type ChildSpec struct {
	Title          string   `json:"title"`
	Kind           string   `json:"kind,omitempty"`
	WorkType       string   `json:"work_type,omitempty"`
	Description    string   `json:"description,omitempty"`
	RequestedActor string   `json:"requested_actor,omitempty"`
	Acceptance     []string `json:"acceptance,omitempty"`
}

// DecomposeProposal records a decomposition proposal for a composite parent: it
// opens a planning run, creates the proposed children in backlog (non-
// dispatchable), moves the parent to review, and creates a pending decompose
// approval carrying the proposed parent contract and child ids (work-tree.md,
// ADR 0009/0030). The children become dispatchable only when the proposal is
// accepted.
func (db *DB) DecomposeProposal(parentID, contractJSON string, children []ChildSpec, actor string) (*Approval, []string, error) {
	if contractJSON == "" {
		contractJSON = "{}"
	}
	if len(children) == 0 {
		return nil, nil, fmt.Errorf("decomposition proposal needs at least one child")
	}
	now := encoding.Now()
	nowT := time.Now()

	var appr *Approval
	var childIDs []string
	err := db.withTx(func(tx *sql.Tx) error {
		parent, err := scanTicket(tx.QueryRow(`SELECT `+ticketColumns+` FROM tickets WHERE id=?`, parentID))
		if err != nil {
			return err
		}
		if parent.NodeType != ticket.NodeComposite {
			return fmt.Errorf("node %s must be triaged composite before decomposition", parentID)
		}

		// Planning run record (records-only in M2; the runtime is Phase 4).
		runID, err := nextSeqID(tx, "run_seq", "R")
		if err != nil {
			return err
		}
		nowStr := encoding.FormatTime(nowT)
		if _, err := tx.Exec(`INSERT INTO runs
			(id, ticket_id, actor_id, actor_snapshot_json, mode, runtime, model, status,
			 workspace_path, started_at, updated_at, completed_at)
			VALUES (?,?,?,?,?,?,?,?,?,?,?,?)`,
			runID, parentID, actor, "{}", string(run.ModePlanning), "stub", nil,
			string(run.StatusCompleted), "", nowStr, nowStr, nowStr); err != nil {
			return err
		}

		// Create proposed children in backlog.
		for _, c := range children {
			id, err := nextTicketID(tx)
			if err != nil {
				return err
			}
			labels, _ := encoding.JSON([]string{})
			acceptance, _ := encoding.JSON(orEmpty(c.Acceptance))
			kind := c.Kind
			if kind == "" {
				kind = "ticket"
			}
			if _, err := tx.Exec(`INSERT INTO tickets
				(id, parent_id, kind, work_type, title, description, contract_json, status,
				 requested_actor, labels_json, acceptance_json, created_at, updated_at)
				VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?)`,
				id, parentID, kind, nullStr(c.WorkType), c.Title, c.Description, "{}",
				string(ticket.StatusBacklog), nullStr(c.RequestedActor), labels, acceptance, now, now); err != nil {
				return err
			}
			childIDs = append(childIDs, id)
			if err := appendAudit(tx, actor, "ticket.created", "ticket", id, map[string]any{
				"parent": parentID, "via": "decompose_proposal",
			}); err != nil {
				return err
			}
		}

		// Move the parent to review (proposal awaiting decision). This is a
		// coordinator-driven state set, like claim/supervise.
		if _, err := tx.Exec(`UPDATE tickets SET status=?, updated_at=? WHERE id=?`,
			string(ticket.StatusReview), now, parentID); err != nil {
			return err
		}
		if err := appendAudit(tx, actor, "ticket.transitioned", "ticket", parentID, map[string]any{
			"to": ticket.StatusReview, "reason": "decompose_proposal",
		}); err != nil {
			return err
		}

		actionJSON, _ := encoding.JSON(map[string]any{"contract": rawJSON(contractJSON), "child_ids": childIDs})
		appr, err = txCreateApproval(tx, CreateApprovalParams{
			RunID: runID, TicketID: parentID, Type: approval.TypeDecompose, RiskClass: string(riskLow),
			Summary:    fmt.Sprintf("Decompose %s into %d children", parentID, len(childIDs)),
			ActionJSON: actionJSON, RequestedByActor: actor,
		})
		return err
	})
	if err != nil {
		return nil, nil, err
	}
	return appr, childIDs, nil
}

// AcceptDecompose approves a decompose proposal: it records the decision, writes
// the proposed contract onto the parent (canon promote-on-dependency, ADR 0013),
// moves the proposed children from backlog to todo, and moves the parent to
// approved. All in one transaction.
func (db *DB) AcceptDecompose(approvalID, decidedBy, reason string) (*Approval, error) {
	var appr *Approval
	err := db.withTx(func(tx *sql.Tx) error {
		a, err := txGetApproval(tx, approvalID)
		if err != nil {
			return err
		}
		if err := requireDecidable(a, approval.TypeDecompose); err != nil {
			return err
		}
		contract, childIDs, err := parseDecomposeAction(a.ActionJSON)
		if err != nil {
			return err
		}
		now := encoding.Now()
		// Promote the parent contract into canon (the forward channel siblings read).
		if _, err := tx.Exec(`UPDATE tickets SET contract_json=?, status=?, updated_at=? WHERE id=?`,
			contract, string(ticket.StatusApproved), now, a.TicketID); err != nil {
			return err
		}
		if err := appendAudit(tx, decidedBy, "ticket.transitioned", "ticket", a.TicketID, map[string]any{
			"to": ticket.StatusApproved, "reason": "decompose_accepted",
		}); err != nil {
			return err
		}
		// Children become dispatchable.
		for _, cid := range childIDs {
			if _, err := tx.Exec(`UPDATE tickets SET status=?, updated_at=? WHERE id=? AND status=?`,
				string(ticket.StatusTodo), now, cid, string(ticket.StatusBacklog)); err != nil {
				return err
			}
			if err := appendAudit(tx, decidedBy, "ticket.transitioned", "ticket", cid, map[string]any{
				"to": ticket.StatusTodo, "reason": "decompose_accepted",
			}); err != nil {
				return err
			}
		}
		if err := txSetApprovalDecision(tx, approvalID, approval.StatusApproved, decidedBy, reason); err != nil {
			return err
		}
		appr, err = txGetApproval(tx, approvalID)
		return err
	})
	if err != nil {
		return nil, err
	}
	return appr, nil
}

// RejectDecompose rejects a decompose proposal: it records the decision and sends
// the parent to rework so the plan can be revised. Proposed children remain in
// backlog (non-dispatchable).
func (db *DB) RejectDecompose(approvalID, decidedBy, reason string) (*Approval, error) {
	var appr *Approval
	err := db.withTx(func(tx *sql.Tx) error {
		a, err := txGetApproval(tx, approvalID)
		if err != nil {
			return err
		}
		if err := requireDecidable(a, approval.TypeDecompose); err != nil {
			return err
		}
		now := encoding.Now()
		if _, err := tx.Exec(`UPDATE tickets SET status=?, updated_at=? WHERE id=?`,
			string(ticket.StatusRework), now, a.TicketID); err != nil {
			return err
		}
		if err := appendAudit(tx, decidedBy, "ticket.transitioned", "ticket", a.TicketID, map[string]any{
			"to": ticket.StatusRework, "reason": "decompose_rejected",
		}); err != nil {
			return err
		}
		if err := txSetApprovalDecision(tx, approvalID, approval.StatusRejected, decidedBy, reason); err != nil {
			return err
		}
		appr, err = txGetApproval(tx, approvalID)
		return err
	})
	if err != nil {
		return nil, err
	}
	return appr, nil
}

const riskLow = "low"

// requireDecidable checks an approval is the expected type and still pending.
func requireDecidable(a *Approval, want approval.Type) error {
	if a.Type != string(want) {
		return fmt.Errorf("approval %s is %s, not %s", a.ID, a.Type, want)
	}
	if approval.Status(a.Status) != approval.StatusPending {
		return fmt.Errorf("%w: approval already %s", ErrIllegalTransition, a.Status)
	}
	return nil
}

// parseDecomposeAction extracts the contract JSON and child ids from a decompose
// approval's action_json.
func parseDecomposeAction(actionJSON string) (contract string, childIDs []string, err error) {
	var parsed struct {
		Contract any      `json:"contract"`
		ChildIDs []string `json:"child_ids"`
	}
	if err := jsonUnmarshal(actionJSON, &parsed); err != nil {
		return "", nil, err
	}
	c, err := encoding.JSON(parsed.Contract)
	if err != nil {
		return "", nil, err
	}
	return c, parsed.ChildIDs, nil
}

// rawJSON wraps a pre-encoded JSON string so it round-trips through encoding.JSON
// as a nested object rather than a quoted string.
func rawJSON(s string) any {
	var v any
	if err := jsonUnmarshal(s, &v); err != nil {
		return map[string]any{}
	}
	return v
}

func jsonUnmarshal(s string, v any) error {
	if s == "" {
		s = "{}"
	}
	return json.Unmarshal([]byte(s), v)
}
