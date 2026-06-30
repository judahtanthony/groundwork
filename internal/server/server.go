// Package server implements gw server, the localhost-only coordinator HTTP
// surface (ADR 0025). It uses the standard-library net/http ServeMux (Go 1.22+
// method+pattern routing) and shares the JSON error envelope with the CLI
// (docs/contracts/http-api.md). Handlers are thin: they call store/coordinator
// services and encode the result; business logic lives elsewhere (overview.md).
package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"sync"
	"time"

	"groundwork/internal/actor"
	"groundwork/internal/config"
	"groundwork/internal/contextbrief"
	"groundwork/internal/eventbus"
	"groundwork/internal/git"
	runpkg "groundwork/internal/run"
	"groundwork/internal/store/sqlite"
	"groundwork/internal/ticket"
)

// Server holds the coordinator's HTTP dependencies and routing table. It is
// coupled to the concrete store and project: it is the coordinator, and the
// context assembler and actor registry are read directly from them.
type Server struct {
	db        *sqlite.DB
	proj      *config.Project
	version   string
	started   time.Time
	mux       *http.ServeMux
	sched     Dispatcher
	bus       *eventbus.Hub
	approvals *ApprovalService
	repo      *git.Repo  // nil when the project root is not a git work tree
	repoMu    sync.Mutex // serializes mutations of the shared main working tree
}

// New builds a coordinator server over the given store and project. version is
// reported by the state endpoint for diagnostics. When the project root is a git
// work tree, landings are committed there (ADR 0034); otherwise the server still
// records landings in the store and skips the commit.
func New(db *sqlite.DB, proj *config.Project, version string) *Server {
	s := &Server{
		db:      db,
		proj:    proj,
		version: version,
		started: time.Now(),
		mux:     http.NewServeMux(),
	}
	if repo, err := git.Open(proj.Root); err == nil {
		s.repo = repo
	}
	s.routes()
	return s
}

// routes registers the endpoint patterns. Method-qualified patterns require the
// Go 1.22+ ServeMux (ADR 0025).
func (s *Server) routes() {
	s.mux.HandleFunc("GET /healthz", s.handleHealth)
	s.mux.HandleFunc("GET /{$}", s.handleDashboard)
	s.mux.HandleFunc("GET /tickets", s.handleTicketsPage)
	s.mux.HandleFunc("GET /approvals", s.handleApprovalsPage)
	s.mux.HandleFunc("GET /review/{id}", s.handleReviewPage)
	s.mux.HandleFunc("POST /approvals/{id}/decide", s.handleApprovalDecideForm)
	s.mux.HandleFunc("GET /static/groundwork.css", s.handleDashboardCSS)
	s.mux.HandleFunc("GET /api/v1/state", s.handleState)
	s.mux.HandleFunc("GET /api/v1/tickets", s.handleTicketList)
	s.mux.HandleFunc("POST /api/v1/tickets", s.handleTicketCreate)
	s.mux.HandleFunc("GET /api/v1/tickets/{id}", s.handleTicketGet)
	s.mux.HandleFunc("PATCH /api/v1/tickets/{id}", s.handleTicketPatch)
	s.mux.HandleFunc("GET /api/v1/tickets/{id}/children", s.handleTicketChildren)
	s.mux.HandleFunc("GET /api/v1/tickets/{id}/context", s.handleTicketContext)
	s.mux.HandleFunc("POST /api/v1/tickets/{id}/transition", s.handleTicketTransition)
	s.mux.HandleFunc("GET /api/v1/tickets/{id}/dependencies", s.handleTicketDependencies)
	s.mux.HandleFunc("POST /api/v1/tickets/{id}/dependencies", s.handleTicketAddDependency)
	s.mux.HandleFunc("DELETE /api/v1/tickets/{id}/dependencies/{depId}", s.handleTicketRemoveDependency)
	s.mux.HandleFunc("POST /api/v1/tickets/{id}/reparent", s.handleTicketReparent)
	s.mux.HandleFunc("GET /api/v1/actors", s.handleActorList)
	s.mux.HandleFunc("GET /api/v1/actors/{id}", s.handleActorGet)
	s.mux.HandleFunc("GET /api/v1/runs", s.handleRunList)
	s.mux.HandleFunc("POST /api/v1/runs", s.handleRunCreate)
	s.mux.HandleFunc("GET /api/v1/runs/{id}", s.handleRunGet)
	s.mux.HandleFunc("GET /api/v1/runs/{id}/events", s.handleRunEvents)
	s.mux.HandleFunc("POST /api/v1/runs/{id}/pause", s.handleRunPause)
	s.mux.HandleFunc("POST /api/v1/runs/{id}/resume", s.handleRunResume)
	s.mux.HandleFunc("POST /api/v1/runs/{id}/cancel", s.handleRunCancel)
	s.mux.HandleFunc("POST /api/v1/tickets/{id}/decompose", s.handleTicketDecompose)
	s.mux.HandleFunc("POST /api/v1/tickets/{id}/escalate", s.handleTicketEscalate)
	s.mux.HandleFunc("POST /api/v1/tickets/{id}/decision", s.handleTicketRaiseDecision)
	s.mux.HandleFunc("POST /api/v1/tickets/{id}/input", s.handleTicketRequestInput)
	s.mux.HandleFunc("GET /api/v1/tickets/{id}/resume", s.handleTicketResume)
	s.mux.HandleFunc("GET /api/v1/tickets/{id}/validations", s.handleTicketValidations)
	s.mux.HandleFunc("POST /api/v1/tickets/{id}/validations", s.handleRecordValidation)
	s.mux.HandleFunc("GET /api/v1/tickets/{id}/land/preview", s.handleTicketLandPreview)
	s.mux.HandleFunc("POST /api/v1/tickets/{id}/land", s.handleTicketLand)
	s.mux.HandleFunc("POST /api/v1/tickets/{id}/land-to-parent", s.handleTicketLandToParent)
	s.mux.HandleFunc("POST /api/v1/tickets/{id}/envelope", s.handleEnvelopePropose)
	s.mux.HandleFunc("GET /api/v1/approvals", s.handleApprovalList)
	s.mux.HandleFunc("GET /api/v1/approvals/{id}", s.handleApprovalGet)
	s.mux.HandleFunc("POST /api/v1/approvals/{id}/approve", s.handleApprovalApprove)
	s.mux.HandleFunc("POST /api/v1/approvals/{id}/reject", s.handleApprovalReject)
	s.mux.HandleFunc("POST /api/v1/approvals/{id}/clarify", s.handleApprovalClarify)
	s.mux.HandleFunc("GET /api/v1/events", s.handleEvents)
}

// SetScheduler attaches the scheduler so run-trigger endpoints (gw run once/next)
// can dispatch work. Read and pause/resume/cancel endpoints work without it.
func (s *Server) SetScheduler(sched Dispatcher) { s.sched = sched }

// SetBus attaches the coordinator event hub used by the SSE endpoint (Wave 5).
func (s *Server) SetBus(bus *eventbus.Hub) { s.bus = bus }

// SetApprovals attaches the approval service used by the decision endpoints.
func (s *Server) SetApprovals(a *ApprovalService) { s.approvals = a }

// Dispatcher is the scheduler surface the server needs to trigger runs.
type Dispatcher interface {
	RunOnce(ctx context.Context, ticketID string) (*sqlite.Run, error)
	Tick(ctx context.Context) (int, error)
}

// Handler exposes the router for testing and embedding.
func (s *Server) Handler() http.Handler { return s.mux }

// Serve binds addr and serves until ctx is cancelled, then shuts down
// gracefully. It writes the bound address to logw once listening so callers can
// report it (and tests can read an ephemeral :0 port). A clean shutdown returns
// nil rather than http.ErrServerClosed.
func (s *Server) Serve(ctx context.Context, addr string, logw io.Writer) error {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	fmt.Fprintf(logw, "gw server listening on http://%s\n", ln.Addr())

	hs := &http.Server{Handler: s.mux}
	errCh := make(chan error, 1)
	go func() { errCh <- hs.Serve(ln) }()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return hs.Shutdown(shutdownCtx)
	}
}

// handleHealth reports store availability (work-tree T-0401). It returns 503
// when the store cannot be reached so external probes see an unhealthy server.
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	if err := s.db.Ping(); err != nil {
		writeError(w, http.StatusServiceUnavailable, "store_unavailable", err.Error())
		return
	}
	// root lets a CLI verify the coordinator serves the same project before
	// routing a mutation to it (T-1033).
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "store": "available", "root": s.proj.Root})
}

// stateResponse is the GET /api/v1/state payload: a coordinator snapshot of
// node counts and server liveness.
type stateResponse struct {
	OK            bool           `json:"ok"`
	Version       string         `json:"version"`
	UptimeSeconds int64          `json:"uptime_seconds"`
	Total         int            `json:"total"`
	Counts        map[string]int `json:"counts"`
	Eligible      int            `json:"eligible"`
}

// handleState returns the current coordinator state snapshot.
func (s *Server) handleState(w http.ResponseWriter, r *http.Request) {
	all, err := s.db.ListTickets()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "list_failed", err.Error())
		return
	}
	eligible, err := s.db.ListEligible()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "list_failed", err.Error())
		return
	}

	counts := map[string]int{}
	for _, t := range all {
		counts[string(t.Status)]++
	}

	writeJSON(w, http.StatusOK, stateResponse{
		OK:            true,
		Version:       s.version,
		UptimeSeconds: int64(time.Since(s.started).Seconds()),
		Total:         len(all),
		Counts:        counts,
		Eligible:      len(eligible),
	})
}

// handleTicketList returns all work nodes ordered by id.
func (s *Server) handleTicketList(w http.ResponseWriter, r *http.Request) {
	all, err := s.db.ListTickets()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "list_failed", err.Error())
		return
	}
	if all == nil {
		all = []*ticket.Ticket{}
	}
	writeJSON(w, http.StatusOK, all)
}

// handleTicketGet returns one work node.
func (s *Server) handleTicketGet(w http.ResponseWriter, r *http.Request) {
	t, err := s.db.GetTicket(r.PathValue("id"))
	if err != nil {
		s.writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, t)
}

// handleTicketCreate creates a work node from the request body and returns it
// (with assigned id and timestamps) with 201.
func (s *Server) handleTicketCreate(w http.ResponseWriter, r *http.Request) {
	var t ticket.Ticket
	if !decodeJSON(w, r, &t) {
		return
	}
	if err := s.db.CreateTicket(&t, ownerActor); err != nil {
		s.writeMutationError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, &t)
}

// handleTicketPatch replaces the mutable fields of a node. The body is the full
// updated representation; status and parentage are not changed here (they move
// through transition; ADR 0022).
func (s *Server) handleTicketPatch(w http.ResponseWriter, r *http.Request) {
	var t ticket.Ticket
	if !decodeJSON(w, r, &t) {
		return
	}
	t.ID = r.PathValue("id")
	if err := s.db.UpdateTicket(&t, ownerActor); err != nil {
		s.writeMutationError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, &t)
}

// handleTicketTransition changes a node's status.
func (s *Server) handleTicketTransition(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Status string `json:"status"`
	}
	if !decodeJSON(w, r, &body) {
		return
	}
	id := r.PathValue("id")
	if err := s.db.TransitionTicket(id, ticket.Status(body.Status), ownerActor); err != nil {
		s.writeMutationError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"id": id, "status": body.Status})
}

// handleTicketAddDependency records that the path node depends on body.depends_on.
func (s *Server) handleTicketAddDependency(w http.ResponseWriter, r *http.Request) {
	var body struct {
		DependsOn string `json:"depends_on"`
	}
	if !decodeJSON(w, r, &body) {
		return
	}
	id := r.PathValue("id")
	if err := s.db.AddDependency(id, body.DependsOn, ownerActor); err != nil {
		s.writeMutationError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"id": id, "depends_on": body.DependsOn, "added": true})
}

// handleTicketRemoveDependency deletes the edge id -> depId.
func (s *Server) handleTicketRemoveDependency(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	depID := r.PathValue("depId")
	if err := s.db.RemoveDependency(id, depID, ownerActor); err != nil {
		s.writeMutationError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"id": id, "depends_on": depID, "added": false})
}

// handleTicketReparent moves the path node under body.parent.
func (s *Server) handleTicketReparent(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Parent string `json:"parent"`
	}
	if !decodeJSON(w, r, &body) {
		return
	}
	id := r.PathValue("id")
	if err := s.db.Reparent(id, body.Parent, ownerActor); err != nil {
		s.writeMutationError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"id": id, "parent": body.Parent, "reparented": true})
}

// handleTicketChildren returns the direct children of a node.
func (s *Server) handleTicketChildren(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if _, err := s.db.GetTicket(id); err != nil {
		s.writeStoreError(w, err)
		return
	}
	children, err := s.db.ListChildren(id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "list_failed", err.Error())
		return
	}
	if children == nil {
		children = []*ticket.Ticket{}
	}
	writeJSON(w, http.StatusOK, children)
}

// handleTicketContext returns the bounded context brief (ADR 0013). The
// ?siblings=true query mirrors the CLI's explicit, non-default sibling query.
func (s *Server) handleTicketContext(w http.ResponseWriter, r *http.Request) {
	brief, err := contextbrief.Build(s.db, s.proj, r.PathValue("id"), r.URL.Query().Get("siblings") == "true")
	if err != nil {
		s.writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, brief)
}

// dependenciesResponse lists both edge directions for a node.
type dependenciesResponse struct {
	DependsOn  []string `json:"depends_on"`
	Dependents []string `json:"dependents"`
}

// handleTicketDependencies returns a node's dependency edges in both directions.
func (s *Server) handleTicketDependencies(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if _, err := s.db.GetTicket(id); err != nil {
		s.writeStoreError(w, err)
		return
	}
	dependsOn, err := s.db.DependencyIDs(id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "list_failed", err.Error())
		return
	}
	dependents, err := s.db.DependentIDs(id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "list_failed", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, dependenciesResponse{DependsOn: dependsOn, Dependents: dependents})
}

// handleActorList exposes the current local actor registry (actors.md). The
// registry is read from .groundwork/actors.yaml on each request so it reflects
// edits without a server restart.
func (s *Server) handleActorList(w http.ResponseWriter, r *http.Request) {
	reg, err := s.loadActors(w)
	if err != nil {
		return
	}
	writeJSON(w, http.StatusOK, reg.Actors)
}

// handleActorGet returns one actor from the registry.
func (s *Server) handleActorGet(w http.ResponseWriter, r *http.Request) {
	reg, err := s.loadActors(w)
	if err != nil {
		return
	}
	a, ok := reg.Get(r.PathValue("id"))
	if !ok {
		writeError(w, http.StatusNotFound, "not_found", fmt.Sprintf("actor %q not found", r.PathValue("id")))
		return
	}
	writeJSON(w, http.StatusOK, a)
}

// loadActors parses the actor registry, writing an error envelope and returning
// a non-nil error when it cannot be read.
func (s *Server) loadActors(w http.ResponseWriter) (*actor.Registry, error) {
	reg, _, err := actor.Load(s.proj.ActorsPath())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "actors_unavailable", err.Error())
		return nil, err
	}
	return reg, nil
}

// ownerActor is the audit actor for API-initiated mutations. The coordinator
// acts on behalf of the local owner in single-user v1 (ADR 0005/0023); chat and
// reviewer-agent decision adapters are later phases.
const ownerActor = "human.owner"

// decodeJSON decodes the request body into v, writing a 400 envelope and
// returning false on malformed input.
func decodeJSON(w http.ResponseWriter, r *http.Request, v any) bool {
	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_body", err.Error())
		return false
	}
	return true
}

// writeStoreError maps store errors to HTTP: ErrNotFound -> 404, else 500.
func (s *Server) writeStoreError(w http.ResponseWriter, err error) {
	if errors.Is(err, sqlite.ErrNotFound) {
		writeError(w, http.StatusNotFound, "not_found", err.Error())
		return
	}
	writeError(w, http.StatusInternalServerError, "store_error", err.Error())
}

// writeMutationError maps store mutation errors to HTTP status + stable codes.
// The codes match what internal/client maps back to store sentinels, so CLI
// behavior is identical over the API and the direct store (ADR 0031).
func (s *Server) writeMutationError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, sqlite.ErrEmptyTitle):
		writeError(w, http.StatusBadRequest, "empty_title", err.Error())
	case errors.Is(err, sqlite.ErrNotFound):
		writeError(w, http.StatusNotFound, "not_found", err.Error())
	case errors.Is(err, sqlite.ErrIllegalTransition):
		writeError(w, http.StatusConflict, "illegal_transition", err.Error())
	case errors.Is(err, sqlite.ErrSelfDependency):
		writeError(w, http.StatusBadRequest, "self_dependency", err.Error())
	case errors.Is(err, sqlite.ErrDependencyCycle):
		writeError(w, http.StatusConflict, "dependency_cycle", err.Error())
	case errors.Is(err, sqlite.ErrSelfParent):
		writeError(w, http.StatusBadRequest, "self_parent", err.Error())
	case errors.Is(err, sqlite.ErrParentCycle):
		writeError(w, http.StatusConflict, "parent_cycle", err.Error())
	case errors.Is(err, sqlite.ErrValidationGate):
		writeError(w, http.StatusConflict, "validation_gate", err.Error())
	case errors.Is(err, sqlite.ErrNotApproved):
		writeError(w, http.StatusConflict, "not_approved", err.Error())
	default:
		writeError(w, http.StatusInternalServerError, "store_error", err.Error())
	}
}

// handleRunList returns runs newest-first.
func (s *Server) handleRunList(w http.ResponseWriter, r *http.Request) {
	runs, err := s.db.ListRuns()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "list_failed", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, runs)
}

// handleRunGet returns one run.
func (s *Server) handleRunGet(w http.ResponseWriter, r *http.Request) {
	run, err := s.db.GetRun(r.PathValue("id"))
	if err != nil {
		s.writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, run)
}

// handleRunEvents returns a run's event log.
func (s *Server) handleRunEvents(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if _, err := s.db.GetRun(id); err != nil {
		s.writeStoreError(w, err)
		return
	}
	events, err := s.db.ListRunEvents(id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "list_failed", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, events)
}

// handleRunCreate triggers a scheduling attempt: with {"ticket_id":...} it runs
// that node once (gw run once); with no body it runs the next eligible node
// (gw run next). Requires the scheduler.
func (s *Server) handleRunCreate(w http.ResponseWriter, r *http.Request) {
	if s.sched == nil {
		writeError(w, http.StatusServiceUnavailable, "scheduler_unavailable", "coordinator scheduler is not running")
		return
	}
	var body struct {
		TicketID string `json:"ticket_id"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body) // empty body is allowed (run next)

	if body.TicketID != "" {
		run, err := s.sched.RunOnce(r.Context(), body.TicketID)
		if err != nil {
			s.writeStoreError(w, err)
			return
		}
		writeJSON(w, http.StatusCreated, run)
		return
	}
	n, err := s.sched.Tick(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "tick_failed", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]int{"started": n})
}

// handleRunPause / Resume / Cancel are the live run-control transitions.
func (s *Server) handleRunPause(w http.ResponseWriter, r *http.Request) {
	s.runStatusChange(w, r, runPause)
}

func (s *Server) handleRunResume(w http.ResponseWriter, r *http.Request) {
	s.runStatusChange(w, r, runResume)
}

func (s *Server) handleRunCancel(w http.ResponseWriter, r *http.Request) {
	s.runStatusChange(w, r, runCancel)
}

type runOp int

const (
	runPause runOp = iota
	runResume
	runCancel
)

func (s *Server) runStatusChange(w http.ResponseWriter, r *http.Request, op runOp) {
	id := r.PathValue("id")
	var err error
	switch op {
	case runPause:
		err = s.db.SetRunStatus(id, runpkg.StatusPaused, ownerActor)
	case runResume:
		err = s.db.SetRunStatus(id, runpkg.StatusRunning, ownerActor)
	case runCancel:
		err = s.db.CancelRun(id, ownerActor)
	}
	if err != nil {
		s.writeMutationError(w, err)
		return
	}
	run, err := s.db.GetRun(id)
	if err != nil {
		s.writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, run)
}

// writeJSON encodes v as indented JSON with the given status code.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	_ = enc.Encode(v)
}

// errorEnvelope is the shared {"error":{code,message}} shape from
// docs/contracts/http-api.md.
type errorEnvelope struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

// writeError writes the JSON error envelope with the given HTTP status.
func writeError(w http.ResponseWriter, status int, code, message string) {
	var e errorEnvelope
	e.Error.Code = code
	e.Error.Message = message
	writeJSON(w, status, e)
}
