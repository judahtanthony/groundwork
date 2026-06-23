package sqlite

import (
	"errors"
	"testing"

	"groundwork/internal/ticket"
)

// mkNode is a small helper to create a node with an explicit id+parent.
func mkNode(t *testing.T, db *DB, id, parent string) {
	t.Helper()
	tk := &ticket.Ticket{ID: id, Kind: "ticket", Title: id, Status: ticket.StatusTodo, ParentID: parent}
	if err := db.CreateTicket(tk, "t"); err != nil {
		t.Fatalf("create %s: %v", id, err)
	}
}

func TestReparentMovesNodeAndAudits(t *testing.T) {
	db := openTestDB(t)
	mkNode(t, db, "T-0001", "")       // root A
	mkNode(t, db, "T-0002", "")       // root B
	mkNode(t, db, "T-0003", "T-0001") // child of A

	if err := db.Reparent("T-0003", "T-0002", "t"); err != nil {
		t.Fatalf("Reparent: %v", err)
	}
	got, err := db.GetTicket("T-0003")
	if err != nil {
		t.Fatal(err)
	}
	if got.ParentID != "T-0002" {
		t.Errorf("parent = %q, want T-0002", got.ParentID)
	}
	// New parent's children include the moved node; old parent's do not.
	kids, _ := db.ListChildren("T-0002")
	if len(kids) != 1 || kids[0].ID != "T-0003" {
		t.Errorf("T-0002 children = %v, want [T-0003]", kids)
	}
	if old, _ := db.ListChildren("T-0001"); len(old) != 0 {
		t.Errorf("T-0001 should have no children, got %v", old)
	}
	// A ticket.reparented audit event was recorded.
	if !hasAuditEvent(t, db, "T-0003", "ticket.reparented") {
		t.Error("missing ticket.reparented audit event")
	}
}

func TestReparentRejectsCycleSelfAndMissing(t *testing.T) {
	db := openTestDB(t)
	mkNode(t, db, "T-0001", "")       // root
	mkNode(t, db, "T-0002", "T-0001") // child
	mkNode(t, db, "T-0003", "T-0002") // grandchild

	// Under its own descendant: cycle.
	if err := db.Reparent("T-0001", "T-0003", "t"); !errors.Is(err, ErrParentCycle) {
		t.Errorf("want ErrParentCycle, got %v", err)
	}
	// Under itself.
	if err := db.Reparent("T-0001", "T-0001", "t"); !errors.Is(err, ErrSelfParent) {
		t.Errorf("want ErrSelfParent, got %v", err)
	}
	// Missing target parent.
	if err := db.Reparent("T-0003", "T-9999", "t"); !errors.Is(err, ErrNotFound) {
		t.Errorf("want ErrNotFound, got %v", err)
	}
	// The rejected moves left the tree unchanged.
	if got, _ := db.GetTicket("T-0001"); got.ParentID != "" {
		t.Errorf("T-0001 parent = %q, want root", got.ParentID)
	}
}

// hasAuditEvent reports whether an audit row of eventType exists for objectID.
func hasAuditEvent(t *testing.T, db *DB, objectID, eventType string) bool {
	t.Helper()
	var n int
	err := db.QueryRow(
		`SELECT COUNT(*) FROM audit_events WHERE object_id=? AND type=?`,
		objectID, eventType).Scan(&n)
	if err != nil {
		t.Fatalf("audit query: %v", err)
	}
	return n > 0
}
