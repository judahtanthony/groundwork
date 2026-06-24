package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"groundwork/internal/ticket"
)

func renderDashboard(t *testing.T, srv *Server) *httptest.ResponseRecorder {
	t.Helper()
	rr := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rr, httptest.NewRequest("GET", "/", nil))
	return rr
}

func TestDashboardRendersHTML(t *testing.T) {
	srv, db := newTestServer(t)
	// A ticket gives the dashboard real content: a status count and an audit event.
	tk := &ticket.Ticket{Title: "Render me", Status: ticket.StatusTodo, WorkType: "technical_implementation"}
	if err := db.CreateTicket(tk, "human.owner"); err != nil {
		t.Fatal(err)
	}

	rr := renderDashboard(t, srv)
	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200\n%s", rr.Code, rr.Body.String())
	}
	if ct := rr.Header().Get("Content-Type"); !strings.HasPrefix(ct, "text/html") {
		t.Errorf("content-type = %q, want text/html", ct)
	}
	body := rr.Body.String()
	for _, want := range []string{
		"Groundwork", "Dashboard",
		"Active runs", "Blocked", "Ready", "Pending approvals", // KPI labels
		"Attention queue", "Recent events", "Work tree", "Local runtime", // panels
		"created", tk.ID, // the ticket's audit event in the timeline
		"No active runs", // empty state for runs
	} {
		if !strings.Contains(body, want) {
			t.Errorf("dashboard HTML missing %q", want)
		}
	}
}

func TestDashboardServesCSS(t *testing.T) {
	srv, _ := newTestServer(t)
	rr := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rr, httptest.NewRequest("GET", "/static/groundwork.css", nil))
	if rr.Code != http.StatusOK {
		t.Fatalf("css status = %d, want 200", rr.Code)
	}
	if ct := rr.Header().Get("Content-Type"); !strings.HasPrefix(ct, "text/css") {
		t.Errorf("css content-type = %q, want text/css", ct)
	}
	if !strings.Contains(rr.Body.String(), "--gw-accent") {
		t.Error("css body missing design tokens")
	}
}

func TestDashboardRootIsExactMatch(t *testing.T) {
	srv, _ := newTestServer(t)
	// "GET /{$}" must match only the exact root, not act as a catch-all.
	rr := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rr, httptest.NewRequest("GET", "/not-a-route", nil))
	if rr.Code != http.StatusNotFound {
		t.Errorf("unknown path status = %d, want 404", rr.Code)
	}
}
