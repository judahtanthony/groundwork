package sqlite

import (
	"fmt"

	"groundwork/internal/decision"
	"groundwork/internal/encoding"
	"groundwork/internal/ticket"
)

// RaiseDecisionParams configures a consequential decision raised by a blocked
// run (ADR 0052). The answer has independent scope/ownership, so it becomes a
// real work node the scheduler routes by work type, with a dependency edge from
// the blocked ticket and a durable decision_requested record explaining the block.
type RaiseDecisionParams struct {
	BlockedTicketID string   // the ticket that is blocked on this decision
	RunID           string   // optional originating run
	Title           string   // decision node title
	WorkType        string   // routes the decision (e.g. architecture_decision)
	RequestedActor  string   // optional actor routing hint
	Statement       string   // the question
	Acceptance      []string // decision node acceptance criteria
	RequestedBy     string   // actor raising the decision
	Parent          string   // optional; defaults to the blocked ticket's parent
}

// RaiseDecision creates a decision work node, links the blocked ticket to it,
// records a durable decision_requested record on the blocked ticket, and moves
// the blocked ticket to blocked — the consequential branch of the ADR 0052
// ladder. All four artifacts are durable before the run is considered safely
// blocked. Returns the new decision node id and the durable record.
//
// It composes existing durable store methods rather than introducing a separate
// decision subsystem (ADR 0052): the decision node is an ordinary work node and
// flows through the same scheduler, policy, SOP, and validation paths.
func (db *DB) RaiseDecision(p RaiseDecisionParams) (string, decision.Record, error) {
	if p.Statement == "" {
		return "", decision.Record{}, fmt.Errorf("raise decision: statement is required")
	}
	if p.WorkType == "" {
		return "", decision.Record{}, fmt.Errorf("raise decision: work_type is required to route the decision")
	}
	blocked, err := db.GetTicket(p.BlockedTicketID)
	if err != nil {
		return "", decision.Record{}, err
	}
	parent := p.Parent
	if parent == "" {
		parent = blocked.ParentID
	}
	actor := p.RequestedBy
	if actor == "" {
		actor = "ai.codex.default"
	}

	dt := &ticket.Ticket{
		Kind: "decision", NodeType: ticket.NodeLeaf, WorkType: p.WorkType,
		Title: p.Title, Status: ticket.StatusTodo, ParentID: parent,
		RequestedActor: p.RequestedActor, Acceptance: p.Acceptance,
	}
	if dt.Title == "" {
		dt.Title = "Decision: " + p.Statement
	}
	if err := db.CreateTicket(dt, actor); err != nil {
		return "", decision.Record{}, fmt.Errorf("create decision node: %w", err)
	}
	if err := db.AddDependency(p.BlockedTicketID, dt.ID, actor); err != nil {
		return "", decision.Record{}, fmt.Errorf("link decision dependency: %w", err)
	}
	rec, err := db.AppendDecision(decision.Record{
		EventType: decision.EventDecisionRequested, TicketID: p.BlockedTicketID, RunID: p.RunID,
		WorkType: p.WorkType, Status: decision.StatusPending, RequestedBy: actor,
		RequestedActor: p.RequestedActor, RequestedAt: encoding.Now(), Statement: p.Statement,
		DependsOn: []string{dt.ID},
	})
	if err != nil {
		return "", decision.Record{}, fmt.Errorf("record decision request: %w", err)
	}
	// Move the blocked ticket to blocked when it is in a state that allows it; a
	// backlog node cannot block (it was never started), so leave it untouched.
	if blocked.Status == ticket.StatusInProgress || blocked.Status == ticket.StatusTodo {
		if err := db.TransitionTicket(p.BlockedTicketID, ticket.StatusBlocked, actor); err != nil {
			return "", decision.Record{}, fmt.Errorf("block originating ticket: %w", err)
		}
	}
	return dt.ID, rec, nil
}

// RecordBlockedHandoff writes the durable blocker for an autonomous run that
// exited blocked/escalated/input-required (ADR 0051), so the block survives a
// rebuild and a later run can resume from it. The outcome maps to the durable
// event type; the handoff summary explains the block for the next run.
func (db *DB) RecordBlockedHandoff(ticketID, runID, outcome, statement, handoffSummary string) error {
	eventType := decision.EventDecisionRequested
	if outcome == "input_required" {
		eventType = decision.EventInputRequested
	}
	if statement == "" {
		statement = handoffSummary
	}
	if handoffSummary == "" {
		handoffSummary = "run exited " + outcome
	}
	_, err := db.AppendDecision(decision.Record{
		EventType: eventType, TicketID: ticketID, RunID: runID, Status: decision.StatusPending,
		RequestedBy: "ai.codex.default", RequestedAt: encoding.Now(),
		Statement: statement, HandoffSummary: handoffSummary,
	})
	return err
}

// RequestInput records a bounded local clarification needed only to continue the
// current run (ADR 0052): a durable input_requested record with no work node and
// no dependency edge. This is the small-uncertainty branch of the ladder — it
// must not spawn a ticket. Returns the durable record.
func (db *DB) RequestInput(ticketID, runID, statement, requestedBy string) (decision.Record, error) {
	if statement == "" {
		return decision.Record{}, fmt.Errorf("request input: statement is required")
	}
	if requestedBy == "" {
		requestedBy = "ai.codex.default"
	}
	return db.AppendDecision(decision.Record{
		EventType: decision.EventInputRequested, TicketID: ticketID, RunID: runID,
		Status: decision.StatusPending, RequestedBy: requestedBy, RequestedAt: encoding.Now(),
		Statement: statement,
	})
}
