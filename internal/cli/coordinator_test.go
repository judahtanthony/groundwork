package cli

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"groundwork/internal/config"
	"groundwork/internal/store/sqlite"
	"groundwork/internal/ticket"
)

// projectWithClosedCoordinator writes a minimal project whose configured server
// address points at a closed port, so the coordinator is never reachable.
func projectWithClosedCoordinator(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	gw := filepath.Join(root, config.GroundworkDir)
	if err := os.MkdirAll(gw, 0o755); err != nil {
		t.Fatal(err)
	}
	cfg := "schema: groundwork_config/v1\nserver:\n  addr: 127.0.0.1:1\n"
	if err := os.WriteFile(filepath.Join(gw, "config.yaml"), []byte(cfg), 0o644); err != nil {
		t.Fatal(err)
	}
	return root
}

func TestRequireCoordinatorFailsWhenDown(t *testing.T) {
	ctx, _, _ := newTestCtx()
	ctx.RootFlag = projectWithClosedCoordinator(t)

	_, err := ctx.requireCoordinator()
	var ce *Error
	if !asError(err, &ce) || ce.Code != "coordinator_required" {
		t.Fatalf("err = %v, want coordinator_required", err)
	}
}

// foreignCoordinator serves a /healthz reporting a different project root.
func foreignCoordinator(t *testing.T) string {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/healthz" {
			w.Header().Set("Content-Type", "application/json")
			io.WriteString(w, `{"status":"ok","store":"available","root":"/some/other/project"}`)
			return
		}
		http.Error(w, "the CLI must not reach a foreign coordinator", http.StatusInternalServerError)
	}))
	t.Cleanup(srv.Close)
	return strings.TrimPrefix(srv.URL, "http://")
}

func TestRequireCoordinatorRejectsForeignProject(t *testing.T) {
	ctx, _, _ := newTestCtx()
	ctx.RootFlag = projectWithCoordinatorAt(t, foreignCoordinator(t))

	_, err := ctx.requireCoordinator()
	var ce *Error
	if !asError(err, &ce) || ce.Code != "coordinator_mismatch" {
		t.Fatalf("err = %v, want coordinator_mismatch", err)
	}
}

func TestOpenTicketStoreSkipsForeignCoordinator(t *testing.T) {
	ctx, _, _ := newTestCtx()
	ctx.RootFlag = projectWithCoordinatorAt(t, foreignCoordinator(t))

	store, closeStore, err := ctx.openTicketStore()
	if err != nil {
		t.Fatalf("openTicketStore: %v", err)
	}
	defer closeStore()
	// A mutation must hit the direct store, not the foreign coordinator.
	if _, ok := store.(*sqlite.DB); !ok {
		t.Errorf("store = %T, want *sqlite.DB (direct fallback)", store)
	}
}

// projectWithCoordinatorAt writes a project whose configured server address is addr.
func projectWithCoordinatorAt(t *testing.T, addr string) string {
	t.Helper()
	root := t.TempDir()
	gw := filepath.Join(root, config.GroundworkDir)
	if err := os.MkdirAll(gw, 0o755); err != nil {
		t.Fatal(err)
	}
	cfg := "schema: groundwork_config/v1\nserver:\n  addr: " + addr + "\n"
	if err := os.WriteFile(filepath.Join(gw, "config.yaml"), []byte(cfg), 0o644); err != nil {
		t.Fatal(err)
	}
	return root
}

func TestOpenTicketStoreFallsBackToStore(t *testing.T) {
	ctx, _, _ := newTestCtx()
	ctx.RootFlag = projectWithClosedCoordinator(t)

	store, closeStore, err := ctx.openTicketStore()
	if err != nil {
		t.Fatalf("openTicketStore: %v", err)
	}
	defer closeStore()

	tk := &ticket.Ticket{Title: "local", Status: ticket.StatusTodo}
	if err := store.CreateTicket(tk, ownerActor); err != nil {
		t.Fatalf("create via fallback store: %v", err)
	}
	got, err := store.GetTicket(tk.ID)
	if err != nil {
		t.Fatalf("get via fallback store: %v", err)
	}
	if got.Title != "local" {
		t.Errorf("title = %q, want local", got.Title)
	}
}
