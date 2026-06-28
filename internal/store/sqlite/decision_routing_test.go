package sqlite

import (
	"testing"

	"groundwork/internal/decision"
	"groundwork/internal/ticket"
)

// TestRaiseDecisionCreatesNodeEdgeAndRecord covers the consequential branch of
// the ADR 0052 ladder (T-1057): a decision work node, a dependency edge, a
// durable decision_requested record, and the blocked transition.
func TestRaiseDecisionCreatesNodeEdgeAndRecord(t *testing.T) {
	db := openTestDB(t)
	parent := seedTicketStatus(t, db, "parent epic", ticket.StatusInProgress)
	blocked := &ticket.Ticket{Title: "implement resume packet", NodeType: ticket.NodeLeaf,
		Status: ticket.StatusInProgress, WorkType: "technical_implementation", ParentID: parent}
	if err := db.CreateTicket(blocked, "tester"); err != nil {
		t.Fatal(err)
	}

	decID, rec, err := db.RaiseDecision(RaiseDecisionParams{
		BlockedTicketID: blocked.ID, RunID: "R-1", Title: "Decide async memory authority model",
		WorkType: "architecture_decision", RequestedActor: "ai.architect.high_context",
		Statement: "Where does async memory authority live?", Acceptance: []string{"ADR records the decision."},
		RequestedBy: "ai.codex.default",
	})
	if err != nil {
		t.Fatal(err)
	}

	// Decision node is a normal work node routed by work type.
	dt, err := db.GetTicket(decID)
	if err != nil {
		t.Fatal(err)
	}
	if dt.WorkType != "architecture_decision" || dt.NodeType != ticket.NodeLeaf || dt.Kind != "decision" {
		t.Fatalf("decision node malformed: %+v", dt)
	}
	if dt.ParentID != parent {
		t.Errorf("decision parent = %q, want inherited %q", dt.ParentID, parent)
	}
	if dt.Status != ticket.StatusTodo {
		t.Errorf("decision status = %q, want todo (routable)", dt.Status)
	}

	// Dependency edge: blocked ticket depends on the decision.
	deps, err := db.DependencyIDs(blocked.ID)
	if err != nil || len(deps) != 1 || deps[0] != decID {
		t.Fatalf("deps = %v err=%v, want [%s]", deps, err, decID)
	}

	// Durable decision_requested record on the blocked ticket, pointing at the node.
	if rec.EventType != decision.EventDecisionRequested || rec.Status != decision.StatusPending {
		t.Errorf("record malformed: %+v", rec)
	}
	if len(rec.DependsOn) != 1 || rec.DependsOn[0] != decID {
		t.Errorf("record depends_on = %v, want [%s]", rec.DependsOn, decID)
	}

	// Originating ticket is now blocked with the record as its explainer.
	got, err := db.GetTicket(blocked.ID)
	if err != nil || got.Status != ticket.StatusBlocked {
		t.Fatalf("blocked status = %q err=%v, want blocked", got.Status, err)
	}
}

func TestRaiseDecisionRequiresWorkTypeAndStatement(t *testing.T) {
	db := openTestDB(t)
	id := seedTicketStatus(t, db, "blocked", ticket.StatusInProgress)
	if _, _, err := db.RaiseDecision(RaiseDecisionParams{BlockedTicketID: id, Statement: "q"}); err == nil {
		t.Error("expected error without work_type")
	}
	if _, _, err := db.RaiseDecision(RaiseDecisionParams{BlockedTicketID: id, WorkType: "architecture_decision"}); err == nil {
		t.Error("expected error without statement")
	}
}

// TestRequestInputDoesNotCreateTicket covers the small-uncertainty branch: a
// local input request records a durable record but spawns no work node (T-1057).
func TestRequestInputDoesNotCreateTicket(t *testing.T) {
	db := openTestDB(t)
	id := seedTicketStatus(t, db, "working", ticket.StatusInProgress)
	before, err := db.ListTickets()
	if err != nil {
		t.Fatal(err)
	}

	rec, err := db.RequestInput(id, "R-1", "Which config key holds the timeout?", "ai.codex.default")
	if err != nil {
		t.Fatal(err)
	}
	if rec.EventType != decision.EventInputRequested {
		t.Fatalf("record type = %q, want input_requested", rec.EventType)
	}

	after, err := db.ListTickets()
	if err != nil {
		t.Fatal(err)
	}
	if len(after) != len(before) {
		t.Fatalf("input request created %d new ticket(s); should create none", len(after)-len(before))
	}
	// The originating ticket stays as-is (no dependency edge, not blocked).
	got, err := db.GetTicket(id)
	if err != nil || got.Status != ticket.StatusInProgress {
		t.Fatalf("status = %q err=%v, want in_progress (unchanged)", got.Status, err)
	}
	deps, _ := db.DependencyIDs(id)
	if len(deps) != 0 {
		t.Errorf("input request added dependency edges %v; should add none", deps)
	}
}
