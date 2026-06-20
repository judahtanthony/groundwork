package client

import (
	"errors"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"groundwork/internal/config"
	"groundwork/internal/server"
	"groundwork/internal/store/sqlite"
	"groundwork/internal/ticket"
)

const actorsYAML = `schema: groundwork_actors/v1
actors:
  - id: human.owner
    type: human
    display_name: Owner
`

// newClient starts a real coordinator over a temp DB and returns a client
// pointed at it plus the underlying store.
func newClient(t *testing.T) (*Client, *sqlite.DB) {
	t.Helper()
	root := t.TempDir()
	gw := filepath.Join(root, config.GroundworkDir)
	if err := os.MkdirAll(gw, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(gw, "actors.yaml"), []byte(actorsYAML), 0o644); err != nil {
		t.Fatal(err)
	}
	db, err := sqlite.Open(filepath.Join(gw, "state.sqlite"))
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	if err := db.Migrate(); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	cfg := config.Defaults()
	srv := server.New(db, &config.Project{Root: root, Config: &cfg}, "test")
	ts := httptest.NewServer(srv.Handler())
	t.Cleanup(ts.Close)
	return New(ts.Listener.Addr().String()), db
}

func TestHealthy(t *testing.T) {
	c, _ := newClient(t)
	if !c.Healthy() {
		t.Error("Healthy() = false, want true")
	}
	if New("127.0.0.1:1").Healthy() {
		t.Error("Healthy() against a closed port = true, want false")
	}
}

func TestCreateAndGet(t *testing.T) {
	c, _ := newClient(t)
	tk := &ticket.Ticket{Title: "via client", Status: ticket.StatusTodo}
	if err := c.CreateTicket(tk, "human.owner"); err != nil {
		t.Fatalf("create: %v", err)
	}
	if tk.ID == "" {
		t.Fatal("create did not assign an id")
	}
	got, err := c.GetTicket(tk.ID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.Title != "via client" {
		t.Errorf("title = %q, want via client", got.Title)
	}
}

func TestGetNotFoundMapsToSentinel(t *testing.T) {
	c, _ := newClient(t)
	_, err := c.GetTicket("T-9999")
	if !errors.Is(err, sqlite.ErrNotFound) {
		t.Fatalf("err = %v, want ErrNotFound", err)
	}
}

func TestUpdate(t *testing.T) {
	c, _ := newClient(t)
	tk := &ticket.Ticket{Title: "before", Status: ticket.StatusTodo}
	if err := c.CreateTicket(tk, "human.owner"); err != nil {
		t.Fatalf("create: %v", err)
	}
	tk.Title = "after"
	if err := c.UpdateTicket(tk, "human.owner"); err != nil {
		t.Fatalf("update: %v", err)
	}
	got, _ := c.GetTicket(tk.ID)
	if got.Title != "after" {
		t.Errorf("title = %q, want after", got.Title)
	}
}

func TestTransitionIllegalMapsToSentinel(t *testing.T) {
	c, _ := newClient(t)
	tk := &ticket.Ticket{Title: "a", Status: ticket.StatusTodo}
	if err := c.CreateTicket(tk, "human.owner"); err != nil {
		t.Fatalf("create: %v", err)
	}
	// todo -> done is not a legal direct transition.
	err := c.TransitionTicket(tk.ID, ticket.StatusDone, "human.owner")
	if !errors.Is(err, sqlite.ErrIllegalTransition) {
		t.Fatalf("err = %v, want ErrIllegalTransition", err)
	}
}

func TestDependencyCycleMapsToSentinel(t *testing.T) {
	c, _ := newClient(t)
	a := &ticket.Ticket{Title: "a", Status: ticket.StatusTodo}
	b := &ticket.Ticket{Title: "b", Status: ticket.StatusTodo}
	if err := c.CreateTicket(a, "human.owner"); err != nil {
		t.Fatal(err)
	}
	if err := c.CreateTicket(b, "human.owner"); err != nil {
		t.Fatal(err)
	}
	if err := c.AddDependency(a.ID, b.ID, "human.owner"); err != nil {
		t.Fatalf("add: %v", err)
	}
	err := c.AddDependency(b.ID, a.ID, "human.owner")
	if !errors.Is(err, sqlite.ErrDependencyCycle) {
		t.Fatalf("err = %v, want ErrDependencyCycle", err)
	}
	if err := c.RemoveDependency(a.ID, b.ID, "human.owner"); err != nil {
		t.Fatalf("remove: %v", err)
	}
}
