package sqlite

import (
	"testing"

	"groundwork/internal/ticket"
)

// makeTree builds parent -> {childA, childB} and returns their ids.
func makeTree(t *testing.T, db *DB) (parent, a, b string) {
	t.Helper()
	p := &ticket.Ticket{Title: "parent", NodeType: ticket.NodeComposite}
	if err := db.CreateTicket(p, "human"); err != nil {
		t.Fatal(err)
	}
	ca := &ticket.Ticket{Title: "child a", ParentID: p.ID}
	cb := &ticket.Ticket{Title: "child b", ParentID: p.ID}
	if err := db.CreateTicket(ca, "human"); err != nil {
		t.Fatal(err)
	}
	if err := db.CreateTicket(cb, "human"); err != nil {
		t.Fatal(err)
	}
	return p.ID, ca.ID, cb.ID
}

func TestListChildren(t *testing.T) {
	db := openTestDB(t)
	parent, a, b := makeTree(t, db)

	kids, err := db.ListChildren(parent)
	if err != nil {
		t.Fatal(err)
	}
	if len(kids) != 2 || kids[0].ID != a || kids[1].ID != b {
		t.Fatalf("children = %v, want [%s %s]", ids(kids), a, b)
	}
}

func TestAncestorsRootFirst(t *testing.T) {
	db := openTestDB(t)
	parent, a, _ := makeTree(t, db)

	anc, err := db.Ancestors(a)
	if err != nil {
		t.Fatal(err)
	}
	if len(anc) != 1 || anc[0].ID != parent {
		t.Fatalf("ancestors = %v, want [%s]", ids(anc), parent)
	}
}

func ids(ts []*ticket.Ticket) []string {
	out := make([]string, len(ts))
	for i, t := range ts {
		out[i] = t.ID
	}
	return out
}
