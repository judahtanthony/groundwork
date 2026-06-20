package sqlite

import (
	"sync"
	"testing"
	"time"

	"groundwork/internal/run"
	"groundwork/internal/ticket"
)

func todoTicket(t *testing.T, db *DB, title string) *ticket.Ticket {
	t.Helper()
	tk := &ticket.Ticket{Title: title, Status: ticket.StatusTodo}
	if err := db.CreateTicket(tk, "tester"); err != nil {
		t.Fatalf("create: %v", err)
	}
	return tk
}

func startParams(ticketID string) StartRunParams {
	return StartRunParams{
		TicketID: ticketID, ActorID: "ai.codex.default", Mode: run.ModeImplementation,
		Runtime: "stub", TTL: 90 * time.Second,
	}
}

func TestStartRunClaimsAndRecords(t *testing.T) {
	db := openTestDB(t)
	tk := todoTicket(t, db, "work")

	r, lease, err := db.StartRun(startParams(tk.ID))
	if err != nil {
		t.Fatalf("StartRun: %v", err)
	}
	if r.ID == "" || r.Status != string(run.StatusRunning) {
		t.Fatalf("run = %+v", r)
	}
	if lease.RunID != r.ID {
		t.Errorf("lease.RunID = %s, want %s", lease.RunID, r.ID)
	}
	// The node is now in_progress and no longer eligible.
	got, _ := db.GetTicket(tk.ID)
	if got.Status != ticket.StatusInProgress {
		t.Errorf("ticket status = %s, want in_progress", got.Status)
	}
}

func TestStartRunSingleWinner(t *testing.T) {
	db := openTestDB(t)
	tk := todoTicket(t, db, "contended")

	const n = 8
	var wg sync.WaitGroup
	var mu sync.Mutex
	wins := 0
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func() {
			defer wg.Done()
			if _, _, err := db.StartRun(startParams(tk.ID)); err == nil {
				mu.Lock()
				wins++
				mu.Unlock()
			}
		}()
	}
	wg.Wait()
	if wins != 1 {
		t.Fatalf("got %d winners, want exactly 1", wins)
	}
	runs, _ := db.ListRunsForTicket(tk.ID)
	if len(runs) != 1 {
		t.Errorf("got %d run rows, want 1", len(runs))
	}
}

func TestSetRunStatusTransitions(t *testing.T) {
	db := openTestDB(t)
	tk := todoTicket(t, db, "lifecycle")
	r, _, _ := db.StartRun(startParams(tk.ID))

	if err := db.SetRunStatus(r.ID, run.StatusPaused, "tester"); err != nil {
		t.Fatalf("pause: %v", err)
	}
	if err := db.SetRunStatus(r.ID, run.StatusRunning, "tester"); err != nil {
		t.Fatalf("resume: %v", err)
	}
	if err := db.SetRunStatus(r.ID, run.StatusCompleted, "tester"); err != nil {
		t.Fatalf("complete: %v", err)
	}
	got, _ := db.GetRun(r.ID)
	if got.Status != string(run.StatusCompleted) || got.CompletedAt == "" {
		t.Errorf("run = %+v, want completed with completed_at", got)
	}
	// Illegal transition from terminal.
	if err := db.SetRunStatus(r.ID, run.StatusRunning, "tester"); err == nil {
		t.Error("expected illegal transition from completed")
	}
}

func TestRunEventsAndCheckpoints(t *testing.T) {
	db := openTestDB(t)
	tk := todoTicket(t, db, "events")
	r, _, _ := db.StartRun(startParams(tk.ID))

	if _, err := db.AppendRunEvent(r.ID, "working", "doing the thing", map[string]any{"step": 1}); err != nil {
		t.Fatalf("append: %v", err)
	}
	if _, err := db.RecordCheckpoint(r.ID, "refs/groundwork/runs/"+r.ID); err != nil {
		t.Fatalf("checkpoint: %v", err)
	}
	if _, err := db.SquashCheckpoints(r.ID); err != nil {
		t.Fatalf("squash: %v", err)
	}
	events, err := db.ListRunEvents(r.ID)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(events) != 3 {
		t.Fatalf("got %d events, want 3", len(events))
	}
	if events[1].EventType != "checkpoint" || events[2].EventType != "checkpoints_squashed" {
		t.Errorf("checkpoint events wrong: %+v", events)
	}
	// last_event/last_message reflect the most recent activity with a message.
	got, _ := db.GetRun(r.ID)
	if got.LastEvent != "checkpoints_squashed" {
		t.Errorf("last_event = %q, want checkpoints_squashed", got.LastEvent)
	}
	if got.LastMessage != "doing the thing" {
		t.Errorf("last_message = %q, want preserved from last message-bearing event", got.LastMessage)
	}
}
