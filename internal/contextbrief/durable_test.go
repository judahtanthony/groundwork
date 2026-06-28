package contextbrief

import (
	"testing"

	"groundwork/internal/completion"
	"groundwork/internal/decision"
	"groundwork/internal/store/sqlite"
	"groundwork/internal/ticket"
)

// insertRunWithChanges adds a completed run row carrying a changed-file set, so
// ChangedFilesForNode returns it (for staleness comparison).
func insertRunWithChanges(t *testing.T, db *sqlite.DB, ticketID string, files []string) {
	t.Helper()
	if _, err := db.Exec(`INSERT INTO runs
		(id, ticket_id, actor_id, actor_snapshot_json, mode, runtime, model, status, workspace_path, started_at, updated_at)
		VALUES (?,?,?,?,?,?,?,?,?,?,?)`,
		"R-1", ticketID, "ai.codex.default", "{}", "implementation", "codex", "m", "completed", "", "2026-06-28T10:00:00Z", "2026-06-28T10:00:00Z"); err != nil {
		t.Fatal(err)
	}
	if err := db.SetRunChangedFiles("R-1", files); err != nil {
		t.Fatal(err)
	}
}

// TestBriefIncludesDurableMemory proves the brief carries durable decision records,
// the completion summary, and a staleness signal (T-1056, ADR 0051/0047).
func TestBriefIncludesDurableMemory(t *testing.T) {
	p, db := newProjectStore(t)
	node := &ticket.Ticket{Title: "implement", NodeType: ticket.NodeLeaf, Status: ticket.StatusReview,
		WorkType: "technical_implementation"}
	if err := db.CreateTicket(node, "human"); err != nil {
		t.Fatal(err)
	}
	if _, err := db.AppendDecision(decision.Record{EventType: decision.EventApprovalRequested,
		TicketID: node.ID, Status: decision.StatusAccepted, Statement: "ok"}); err != nil {
		t.Fatal(err)
	}
	if _, err := db.AppendDecision(decision.Record{EventType: decision.EventDecisionRequested,
		TicketID: node.ID, Status: decision.StatusPending, Statement: "which lib?", HandoffSummary: "blocked"}); err != nil {
		t.Fatal(err)
	}
	// Summary recorded one changed file; a later run touched a different file → stale.
	if err := completion.Write(p.TicketsDir(), &completion.Summary{NodeID: node.ID, Outcome: "produced", Changed: []string{"a.go"}}); err != nil {
		t.Fatal(err)
	}
	if err := db.UpsertCompletionSummary(&completion.Summary{NodeID: node.ID, Outcome: "produced", Changed: []string{"a.go"}}); err != nil {
		t.Fatal(err)
	}
	insertRunWithChanges(t, db, node.ID, []string{"b.go"})

	b, err := Build(db, p, node.ID, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(b.PendingBlockers) != 1 || b.PendingBlockers[0].HandoffSummary != "blocked" {
		t.Errorf("pending blockers = %+v", b.PendingBlockers)
	}
	if len(b.RecentDecisions) != 1 {
		t.Errorf("recent decisions = %+v", b.RecentDecisions)
	}
	if b.CompletionSummary == nil {
		t.Fatal("completion summary not included")
	}
	if !b.SummaryStale || b.SummaryStaleReason == "" {
		t.Errorf("expected stale summary signal, got stale=%v", b.SummaryStale)
	}
}

// TestBriefSignalsMissingSummary flags a review node with no summary.
func TestBriefSignalsMissingSummary(t *testing.T) {
	p, db := newProjectStore(t)
	node := &ticket.Ticket{Title: "n", NodeType: ticket.NodeLeaf, Status: ticket.StatusReview, WorkType: "technical_implementation"}
	if err := db.CreateTicket(node, "human"); err != nil {
		t.Fatal(err)
	}
	b, err := Build(db, p, node.ID, false)
	if err != nil {
		t.Fatal(err)
	}
	if !b.SummaryMissing {
		t.Error("expected SummaryMissing for a review node without a summary")
	}
}
