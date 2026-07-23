package server

import (
	"net/http"
	"os"
	"strings"
	"testing"
)

func TestSettingsAndAgentsMDSync(t *testing.T) {
	srv, _ := newTestServer(t)
	var settings settingsResponse
	if code := get(t, srv, "/api/v1/settings", &settings); code != http.StatusOK {
		t.Fatalf("settings status = %d, want 200", code)
	}
	if settings.RepositoryPath != srv.proj.Root || settings.SQLitePath != srv.proj.DBPath() {
		t.Fatalf("paths = %#v", settings)
	}
	if settings.Server.Address != "127.0.0.1:4500" || settings.Server.Bind != "127.0.0.1" || settings.Server.Port != "4500" {
		t.Fatalf("server = %#v", settings.Server)
	}
	if settings.Agent.Engine != "codex" || settings.Agent.Sandbox != "workspace-write" {
		t.Fatalf("agent = %#v", settings.Agent)
	}
	if settings.Concurrency.Max != 4 || settings.Concurrency.LeaseTTL != "1m30s" || settings.Concurrency.LeaseHeartbeat != "30s" {
		t.Fatalf("concurrency = %#v", settings.Concurrency)
	}
	if settings.AgentsMD.State != "missing" {
		t.Fatalf("agents state = %q, want missing", settings.AgentsMD.State)
	}

	var synced map[string]string
	if code := req(t, srv, http.MethodPost, "/api/v1/settings/agents-md/sync", nil, &synced); code != http.StatusOK {
		t.Fatalf("sync status = %d, want 200", code)
	}
	body, err := os.ReadFile(synced["path"])
	if err != nil {
		t.Fatal(err)
	}
	if synced["state"] != "synced" || !strings.Contains(string(body), ".groundwork/WORKFLOW.md") {
		t.Fatalf("sync = %#v\n%s", synced, body)
	}
}

func TestDoctorEndpointUsesCLIHealthChecks(t *testing.T) {
	srv, _ := newTestServer(t)
	var report struct {
		Healthy bool `json:"healthy"`
		Checks  []struct {
			Name string `json:"name"`
		} `json:"checks"`
	}
	if code := req(t, srv, http.MethodPost, "/api/v1/doctor", nil, &report); code != http.StatusOK {
		t.Fatalf("status = %d, want 200", code)
	}
	if !report.Healthy || len(report.Checks) < 4 {
		t.Fatalf("report = %#v", report)
	}
}
