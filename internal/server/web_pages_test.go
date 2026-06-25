package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// getHTML issues a GET against the server and returns the recorder.
func getHTML(t *testing.T, srv *Server, path string) *httptest.ResponseRecorder {
	t.Helper()
	rr := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rr, httptest.NewRequest("GET", path, nil))
	return rr
}

// The dashboard nav must link Tickets and Approvals as live routes (no "soon"
// placeholder), proving the activated navigation of the operator shell.
func TestNavActivatesOperatorPages(t *testing.T) {
	srv, _ := newTestServer(t)
	body := getHTML(t, srv, "/").Body.String()
	for _, want := range []string{`href="/tickets"`, `href="/approvals"`} {
		if !strings.Contains(body, want) {
			t.Errorf("dashboard nav missing activated link %q", want)
		}
	}
}

// Each operator page renders on the shared chrome (sidebar, topbar, SSE script)
// with its own active-nav highlight, so the layout is reused, not duplicated.
func TestOperatorPagesRenderOnSharedChrome(t *testing.T) {
	srv, _ := newTestServer(t)
	cases := []struct {
		path, active, title string
	}{
		{"/tickets", `<a class="active" href="/tickets">`, "Tickets"},
		{"/approvals", `<a class="active" href="/approvals">`, "Approvals"},
	}
	for _, tc := range cases {
		rr := getHTML(t, srv, tc.path)
		if rr.Code != http.StatusOK {
			t.Fatalf("GET %s = %d, want 200\n%s", tc.path, rr.Code, rr.Body.String())
		}
		if ct := rr.Header().Get("Content-Type"); !strings.HasPrefix(ct, "text/html") {
			t.Errorf("GET %s content-type = %q, want text/html", tc.path, ct)
		}
		body := rr.Body.String()
		// Shared chrome present on every page.
		for _, want := range []string{"Groundwork", `id="live"`, "/static/groundwork.css", `href="/"`} {
			if !strings.Contains(body, want) {
				t.Errorf("GET %s missing shared chrome %q", tc.path, want)
			}
		}
		// Active-nav highlight for this page.
		if !strings.Contains(body, tc.active) {
			t.Errorf("GET %s missing active nav %q", tc.path, tc.active)
		}
		// Breadcrumb leaf / title.
		if !strings.Contains(body, tc.title) {
			t.Errorf("GET %s missing title %q", tc.path, tc.title)
		}
	}
}
