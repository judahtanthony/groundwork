package cli

import (
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"

	"groundwork/internal/store/sqlite"
	"groundwork/internal/ticket"
)

// cliTestDB opens a migrated store in a temp dir for CLI read tests.
func cliTestDB(t *testing.T) *sqlite.DB {
	t.Helper()
	db, err := sqlite.Open(filepath.Join(t.TempDir(), "state.sqlite"))
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	if err := db.Migrate(); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return db
}

// readyBlockedFixture builds: a done dependency, an eligible todo node, and a
// todo node blocked by an unmet dependency. It returns their ids.
func readyBlockedFixture(t *testing.T, db *sqlite.DB) (doneID, readyID, blockedID, blockerID string) {
	t.Helper()
	mk := func(title string, status ticket.Status) *ticket.Ticket {
		tk := &ticket.Ticket{Title: title, Status: status, WorkType: "technical_implementation"}
		if err := db.CreateTicket(tk, "t"); err != nil {
			t.Fatalf("create %s: %v", title, err)
		}
		return tk
	}
	done := mk("done dep", ticket.StatusDone)
	ready := mk("ready node", ticket.StatusTodo)
	blocker := mk("pending dep", ticket.StatusTodo)
	blocked := mk("blocked node", ticket.StatusTodo)
	// ready depends on a done node (still eligible); blocked depends on a todo node.
	if err := db.AddDependency(ready.ID, done.ID, "t"); err != nil {
		t.Fatalf("link ready: %v", err)
	}
	if err := db.AddDependency(blocked.ID, blocker.ID, "t"); err != nil {
		t.Fatalf("link blocked: %v", err)
	}
	return done.ID, ready.ID, blocked.ID, blocker.ID
}

func TestListReadyShowsEligibleNotBlocked(t *testing.T) {
	db := cliTestDB(t)
	_, readyID, blockedID, _ := readyBlockedFixture(t, db)

	ctx, out, _ := newTestCtx()
	if err := listReady(ctx, db); err != nil {
		t.Fatalf("listReady: %v", err)
	}
	got := out.String()
	if !strings.Contains(got, readyID) {
		t.Errorf("ready output missing eligible node %s:\n%s", readyID, got)
	}
	if strings.Contains(got, blockedID) {
		t.Errorf("ready output should not list blocked node %s:\n%s", blockedID, got)
	}
}

func TestListBlockedAnnotatesBlockers(t *testing.T) {
	db := cliTestDB(t)
	_, readyID, blockedID, blockerID := readyBlockedFixture(t, db)

	ctx, out, _ := newTestCtx()
	if err := listBlocked(ctx, db); err != nil {
		t.Fatalf("listBlocked: %v", err)
	}
	got := out.String()
	if !strings.Contains(got, blockedID) {
		t.Errorf("blocked output missing %s:\n%s", blockedID, got)
	}
	if !strings.Contains(got, "blocked by: "+blockerID) {
		t.Errorf("blocked output missing blocker annotation %s:\n%s", blockerID, got)
	}
	if strings.Contains(got, readyID) {
		t.Errorf("eligible node %s should not appear in --blocked:\n%s", readyID, got)
	}
}

func TestClaimNodeEligibleSetsAssigneeAndStarts(t *testing.T) {
	db := cliTestDB(t)
	_, readyID, _, _ := readyBlockedFixture(t, db)

	if err := claimNode(db, db, readyID, "human.owner"); err != nil {
		t.Fatalf("claimNode: %v", err)
	}
	tk, err := db.GetTicket(readyID)
	if err != nil {
		t.Fatal(err)
	}
	if tk.Status != ticket.StatusInProgress {
		t.Errorf("status = %s, want in_progress", tk.Status)
	}
	if tk.Assignee != "human.owner" {
		t.Errorf("assignee = %q, want human.owner", tk.Assignee)
	}
}

func TestClaimNodeRefusesBlocked(t *testing.T) {
	db := cliTestDB(t)
	_, _, blockedID, blockerID := readyBlockedFixture(t, db)

	err := claimNode(db, db, blockedID, "human.owner")
	if err == nil {
		t.Fatal("expected claim of a blocked node to fail")
	}
	var ce *Error
	if !asError(err, &ce) || ce.Code != "blocked" {
		t.Fatalf("want blocked error, got %v", err)
	}
	if !strings.Contains(ce.Message, blockerID) {
		t.Errorf("error should name the blocker %s: %q", blockerID, ce.Message)
	}
	// The node must remain todo (not partially mutated).
	tk, _ := db.GetTicket(blockedID)
	if tk.Status != ticket.StatusTodo {
		t.Errorf("blocked node moved to %s; must stay todo", tk.Status)
	}
}

func TestClaimNodeRefusesNonTodo(t *testing.T) {
	db := cliTestDB(t)
	doneID, _, _, _ := readyBlockedFixture(t, db)

	err := claimNode(db, db, doneID, "human.owner")
	if err == nil {
		t.Fatal("expected claim of a done node to fail")
	}
	var ce *Error
	if !asError(err, &ce) || ce.Code != "not_claimable" {
		t.Fatalf("want not_claimable error, got %v", err)
	}
}

func TestListBlockedJSONHasBlockedBy(t *testing.T) {
	db := cliTestDB(t)
	_, _, blockedID, blockerID := readyBlockedFixture(t, db)

	ctx, out, _ := newTestCtx()
	ctx.JSON = true
	if err := listBlocked(ctx, db); err != nil {
		t.Fatalf("listBlocked: %v", err)
	}
	var nodes []blockedNode
	if err := json.Unmarshal(out.Bytes(), &nodes); err != nil {
		t.Fatalf("unmarshal: %v\n%s", err, out.String())
	}
	if len(nodes) != 1 {
		t.Fatalf("want 1 blocked node, got %d", len(nodes))
	}
	if nodes[0].ID != blockedID {
		t.Errorf("blocked id = %s, want %s", nodes[0].ID, blockedID)
	}
	if len(nodes[0].BlockedBy) != 1 || nodes[0].BlockedBy[0].ID != blockerID {
		t.Errorf("blocked_by = %+v, want [%s]", nodes[0].BlockedBy, blockerID)
	}
	if nodes[0].BlockedBy[0].Status != string(ticket.StatusTodo) {
		t.Errorf("blocker status = %s, want todo", nodes[0].BlockedBy[0].Status)
	}
}
