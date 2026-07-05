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

func TestCoordinatorRoot(t *testing.T) {
	c, _ := newClient(t)
	root, ok := c.CoordinatorRoot()
	if !ok {
		t.Error("CoordinatorRoot ok = false, want true")
	}
	if root == "" {
		t.Error("CoordinatorRoot root = empty, want the served project root")
	}
	if _, ok := New("127.0.0.1:1").CoordinatorRoot(); ok {
		t.Error("CoordinatorRoot against a closed port ok = true, want false")
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

// A plain (unparented) ticket lands to main: the land route is "main", so the CLI
// keeps the working-tree staging path rather than routing to land_to_parent.
func TestLandRouteDefaultsToMain(t *testing.T) {
	c, _ := newClient(t)
	tk := &ticket.Ticket{Title: "root", Status: ticket.StatusTodo}
	if err := c.CreateTicket(tk, "human.owner"); err != nil {
		t.Fatal(err)
	}
	route, branch, err := c.LandRoute(tk.ID)
	if err != nil {
		t.Fatalf("LandRoute: %v", err)
	}
	if route != "main" {
		t.Errorf("route = %q, want main", route)
	}
	if branch != "" {
		t.Errorf("branch = %q, want empty for the main route", branch)
	}
}

// LandToParent hits the land-to-parent endpoint and surfaces the server's
// no_integration_target error when the node has no integration branch in its
// chain (mirrors LandTicket's error mapping).
func TestLandToParentWithoutTargetSurfacesError(t *testing.T) {
	c, _ := newClient(t)
	tk := &ticket.Ticket{Title: "orphan", Status: ticket.StatusTodo}
	if err := c.CreateTicket(tk, "human.owner"); err != nil {
		t.Fatal(err)
	}
	_, err := c.LandToParent(tk.ID)
	if err == nil {
		t.Fatal("LandToParent without an integration target: want error")
	}
	var apiErr *APIError
	if !errors.As(err, &apiErr) || apiErr.Code != "no_integration_target" {
		t.Fatalf("err = %v, want APIError code no_integration_target", err)
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
