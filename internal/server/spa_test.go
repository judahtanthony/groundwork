package server

import (
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"
)

var scriptSrc = regexp.MustCompile(`src="([^"]+\.js)"`)

func TestEmbeddedSPAServesIndex(t *testing.T) {
	srv, _ := newTestServer(t)
	rr := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/app/", nil))

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200\n%s", rr.Code, rr.Body.String())
	}
	if ct := rr.Header().Get("Content-Type"); !strings.HasPrefix(ct, "text/html") {
		t.Errorf("content-type = %q, want text/html", ct)
	}
	body := rr.Body.String()
	if !strings.Contains(body, "Groundwork") || !strings.Contains(body, "/app/assets/") {
		t.Errorf("embedded index is not the Vite production build: %s", body)
	}

	match := scriptSrc.FindStringSubmatch(body)
	if len(match) != 2 {
		t.Fatalf("index has no production script asset: %s", body)
	}
	asset := httptest.NewRecorder()
	srv.Handler().ServeHTTP(asset, httptest.NewRequest(http.MethodGet, match[1], nil))
	if asset.Code != http.StatusOK {
		t.Fatalf("asset %s status = %d, want 200", match[1], asset.Code)
	}
	if ct := asset.Header().Get("Content-Type"); !strings.Contains(ct, "javascript") {
		t.Errorf("asset content-type = %q, want javascript", ct)
	}
}

func TestEmbeddedSPARedirectsToTrailingSlash(t *testing.T) {
	srv, _ := newTestServer(t)
	rr := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/app", nil))

	if rr.Code != http.StatusPermanentRedirect {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusPermanentRedirect)
	}
	if got := rr.Header().Get("Location"); got != "/app/" {
		t.Errorf("location = %q, want /app/", got)
	}
}
