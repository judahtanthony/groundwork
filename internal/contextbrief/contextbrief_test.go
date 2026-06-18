package contextbrief

import (
	"path/filepath"
	"testing"

	"groundwork/internal/config"
	"groundwork/internal/store/sqlite"
	"groundwork/internal/ticket"
)

func newProjectStore(t *testing.T) (*config.Project, *sqlite.DB) {
	t.Helper()
	root := t.TempDir()
	cfg := config.Defaults()
	p := &config.Project{Root: root, Config: &cfg}
	db, err := sqlite.Open(filepath.Join(root, "state.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	if err := db.Migrate(); err != nil {
		t.Fatal(err)
	}
	return p, db
}

func TestBuild(t *testing.T) {
	p, db := newProjectStore(t)

	parent := &ticket.Ticket{Title: "Build store", NodeType: ticket.NodeComposite, Contract: `{"schema":"v1"}`}
	if err := db.CreateTicket(parent, "human"); err != nil {
		t.Fatal(err)
	}
	dep := &ticket.Ticket{Title: "Schema", ParentID: parent.ID}
	node := &ticket.Ticket{Title: "Migrations", ParentID: parent.ID}
	sibling := &ticket.Ticket{Title: "CRUD", ParentID: parent.ID}
	for _, tk := range []*ticket.Ticket{dep, node, sibling} {
		if err := db.CreateTicket(tk, "human"); err != nil {
			t.Fatal(err)
		}
	}
	if err := db.AddDependency(node.ID, dep.ID, "human"); err != nil {
		t.Fatal(err)
	}

	b, err := Build(db, p, node.ID, false)
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	if len(b.AncestorSpine) != 1 || b.AncestorSpine[0].ID != parent.ID {
		t.Errorf("ancestor spine = %+v, want [%s]", b.AncestorSpine, parent.ID)
	}
	if b.ParentContract != `{"schema":"v1"}` {
		t.Errorf("parent contract = %q", b.ParentContract)
	}
	if len(b.Dependencies) != 1 || b.Dependencies[0].ID != dep.ID {
		t.Errorf("dependencies = %+v, want [%s]", b.Dependencies, dep.ID)
	}
	if b.Siblings != nil {
		t.Errorf("siblings should be nil without includeSiblings: %+v", b.Siblings)
	}
}

func TestBuildWithSiblings(t *testing.T) {
	p, db := newProjectStore(t)
	parent := &ticket.Ticket{Title: "p", NodeType: ticket.NodeComposite}
	if err := db.CreateTicket(parent, "human"); err != nil {
		t.Fatal(err)
	}
	node := &ticket.Ticket{Title: "n", ParentID: parent.ID}
	sib := &ticket.Ticket{Title: "s", ParentID: parent.ID}
	for _, tk := range []*ticket.Ticket{node, sib} {
		if err := db.CreateTicket(tk, "human"); err != nil {
			t.Fatal(err)
		}
	}

	b, err := Build(db, p, node.ID, true)
	if err != nil {
		t.Fatal(err)
	}
	if len(b.Siblings) != 1 || b.Siblings[0].ID != sib.ID {
		t.Errorf("siblings = %+v, want [%s]", b.Siblings, sib.ID)
	}
}
