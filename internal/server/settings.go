package server

import (
	"net"
	"net/http"

	"groundwork/internal/agentsync"
	"groundwork/internal/doctor"
)

type settingsResponse struct {
	RepositoryPath string           `json:"repository_path"`
	SQLitePath     string           `json:"sqlite_path"`
	ConfigPath     string           `json:"config_path"`
	Server         settingsServer   `json:"server"`
	Agent          settingsAgent    `json:"agent"`
	Concurrency    settingsSchedule `json:"concurrency"`
	AgentsMD       agentsync.Status `json:"agents_md"`
}

type settingsServer struct {
	Address string `json:"address"`
	Bind    string `json:"bind"`
	Port    string `json:"port"`
}

type settingsAgent struct {
	Engine  string `json:"engine"`
	Model   string `json:"model,omitempty"`
	Sandbox string `json:"sandbox"`
}

type settingsSchedule struct {
	Max            int    `json:"max"`
	LeaseTTL       string `json:"lease_ttl"`
	LeaseHeartbeat string `json:"lease_heartbeat"`
}

func (s *Server) handleSettings(w http.ResponseWriter, r *http.Request) {
	syncStatus, err := agentsync.Inspect(s.proj.Root)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "agents_status_failed", err.Error())
		return
	}
	cfg := s.proj.Config
	server := settingsServer{Address: cfg.Server.Addr}
	if host, port, err := net.SplitHostPort(cfg.Server.Addr); err == nil {
		server.Bind = host
		server.Port = port
	}
	writeJSON(w, http.StatusOK, settingsResponse{
		RepositoryPath: s.proj.Root,
		SQLitePath:     s.proj.DBPath(),
		ConfigPath:     s.proj.ConfigPath(),
		Server:         server,
		Agent: settingsAgent{
			Engine: cfg.Runtime, Model: cfg.Model, Sandbox: cfg.Sandbox,
		},
		Concurrency: settingsSchedule{
			Max:            cfg.MaxConcurrency,
			LeaseTTL:       cfg.Lease.TTL.Duration().String(),
			LeaseHeartbeat: cfg.Lease.Heartbeat.Duration().String(),
		},
		AgentsMD: syncStatus,
	})
}

func (s *Server) handleAgentsMDSync(w http.ResponseWriter, r *http.Request) {
	status, err := agentsync.Sync(s.proj.Root)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "agents_sync_failed", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, status)
}

func (s *Server) handleDoctor(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, doctor.Run(s.proj))
}
