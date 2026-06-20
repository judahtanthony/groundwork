package sqlite

import (
	"testing"

	"groundwork/internal/run"
	"groundwork/internal/ticket"
)

func TestReconcileStartupInterruptsAndRequeues(t *testing.T) {
	db := openTestDB(t)
	tk := todoTicket(t, db, "in flight")
	r, _, err := db.StartRun(startParams(tk.ID))
	if err != nil {
		t.Fatalf("StartRun: %v", err)
	}
	// Sanity: node claimed, lease present, run running.
	if got, _ := db.GetTicket(tk.ID); got.Status != ticket.StatusInProgress {
		t.Fatalf("precondition: ticket = %s", got.Status)
	}

	rep, err := db.ReconcileStartup()
	if err != nil {
		t.Fatalf("ReconcileStartup: %v", err)
	}
	if rep.InterruptedRuns != 1 || rep.ReleasedLeases != 1 || rep.RequeuedNodes != 1 {
		t.Fatalf("report = %+v", rep)
	}
	if got, _ := db.GetRun(r.ID); got.Status != string(run.StatusInterrupted) {
		t.Errorf("run status = %s, want interrupted", got.Status)
	}
	if lease, _ := db.GetLease(tk.ID); lease != nil {
		t.Errorf("lease still present: %+v", lease)
	}
	if got, _ := db.GetTicket(tk.ID); got.Status != ticket.StatusTodo {
		t.Errorf("ticket status = %s, want todo (requeued)", got.Status)
	}
}

func TestReconcileLeavesTerminalRuns(t *testing.T) {
	db := openTestDB(t)
	tk := todoTicket(t, db, "done work")
	r, _, _ := db.StartRun(startParams(tk.ID))
	if err := db.SetRunStatus(r.ID, run.StatusCompleted, "tester"); err != nil {
		t.Fatal(err)
	}
	rep, err := db.ReconcileStartup()
	if err != nil {
		t.Fatal(err)
	}
	if rep.InterruptedRuns != 0 {
		t.Errorf("interrupted = %d, want 0 (run already completed)", rep.InterruptedRuns)
	}
	if got, _ := db.GetRun(r.ID); got.Status != string(run.StatusCompleted) {
		t.Errorf("completed run was altered: %s", got.Status)
	}
}
