package resume

import (
	"path/filepath"
	"testing"

	"groundwork/internal/decision"
	"groundwork/internal/store/sqlite"
	"groundwork/internal/ticket"
)

func testDB(t *testing.T) *sqlite.DB {
	t.Helper()
	db, err := sqlite.Open(filepath.Join(t.TempDir(), "state.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	if err := db.Migrate(); err != nil {
		t.Fatal(err)
	}
	return db
}

func TestAssembleBuildsResumePacket(t *testing.T) {
	db := testDB(t)
	parent := &ticket.Ticket{Title: "epic", NodeType: ticket.NodeComposite, Status: ticket.StatusInProgress,
		WorkType: "technical_design", Contract: `{"schema":"contract/v1","goal":"x"}`}
	if err := db.CreateTicket(parent, "tester"); err != nil {
		t.Fatal(err)
	}
	dep := &ticket.Ticket{Title: "dep", NodeType: ticket.NodeLeaf, Status: ticket.StatusDone, WorkType: "technical_implementation"}
	if err := db.CreateTicket(dep, "tester"); err != nil {
		t.Fatal(err)
	}
	node := &ticket.Ticket{ParentID: parent.ID, Title: "implement", NodeType: ticket.NodeLeaf,
		Status: ticket.StatusBlocked, WorkType: "technical_implementation", Acceptance: []string{"tests pass"}}
	if err := db.CreateTicket(node, "tester"); err != nil {
		t.Fatal(err)
	}
	if err := db.AddDependency(node.ID, dep.ID, "tester"); err != nil {
		t.Fatal(err)
	}
	// A resolved decision and a pending blocker with a handoff summary.
	if _, err := db.AppendDecision(decision.Record{ID: "D-1", EventType: decision.EventApprovalRequested,
		TicketID: node.ID, Status: decision.StatusAccepted, Statement: "ok?"}); err != nil {
		t.Fatal(err)
	}
	if _, err := db.AppendDecision(decision.Record{EventType: decision.EventDecisionRequested,
		TicketID: node.ID, Status: decision.StatusPending, Statement: "which API?",
		HandoffSummary: "blocked on the API choice"}); err != nil {
		t.Fatal(err)
	}

	p, err := Assemble(db, node.ID)
	if err != nil {
		t.Fatal(err)
	}
	if p.TicketID != node.ID || p.WorkType != "technical_implementation" {
		t.Fatalf("packet basics wrong: %+v", p)
	}
	if p.AncestorContract == "" {
		t.Error("ancestor contract not assembled")
	}
	if len(p.Dependencies) != 1 || p.Dependencies[0].Status != string(ticket.StatusDone) {
		t.Errorf("dependencies = %+v", p.Dependencies)
	}
	if len(p.PendingBlockers) != 1 || len(p.ResolvedDecisions) != 1 {
		t.Errorf("decision split wrong: pending=%d resolved=%d", len(p.PendingBlockers), len(p.ResolvedDecisions))
	}
	if p.HandoffSummary != "blocked on the API choice" {
		t.Errorf("handoff = %q", p.HandoffSummary)
	}
	if p.NextAction == "" || p.NextAction[:7] != "resolve" {
		t.Errorf("next action = %q, want a resolve-blocker recommendation", p.NextAction)
	}
}

func TestAssembleCleanNodeRecommendsContinue(t *testing.T) {
	db := testDB(t)
	node := &ticket.Ticket{Title: "work", NodeType: ticket.NodeLeaf, Status: ticket.StatusInProgress,
		WorkType: "technical_implementation", Acceptance: []string{"done"}}
	if err := db.CreateTicket(node, "tester"); err != nil {
		t.Fatal(err)
	}
	p, err := Assemble(db, node.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(p.PendingBlockers) != 0 {
		t.Errorf("unexpected blockers: %+v", p.PendingBlockers)
	}
	if p.NextAction != "continue implementation toward the acceptance criteria" {
		t.Errorf("next action = %q", p.NextAction)
	}
}
