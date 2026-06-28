// Package scheduler is the coordinator's serial scheduling loop plus its bounded
// pool of run supervisors (ADR 0026). Each tick finds eligible nodes, selects an
// authorized actor (gate engine + actor capability, ADR 0028/0029), claims the
// node transactionally (single-winner), and hands it to a supervisor goroutine
// that drives the runtime and renews the lease. Claim arbitration is serial;
// execution fans out up to MaxConcurrency.
package scheduler

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"time"

	"groundwork/internal/actor"
	"groundwork/internal/eventbus"
	"groundwork/internal/policy"
	"groundwork/internal/risk"
	"groundwork/internal/run"
	"groundwork/internal/runtime"
	"groundwork/internal/store/sqlite"
	"groundwork/internal/ticket"
)

// Config tunes the loop (defaults from docs/architecture/runtime-model.md).
type Config struct {
	MaxConcurrency int
	LeaseTTL       time.Duration
	Heartbeat      time.Duration
	TickInterval   time.Duration
	Model          string
	RunLogDir      string // per-run events.ndjson root (.groundwork/runs); "" disables
}

// Scheduler owns scheduling decisions and supervises runs.
type Scheduler struct {
	db       *sqlite.DB
	policies *policy.Set
	registry *actor.Registry
	rt       runtime.Runtime
	bus      *eventbus.Hub
	cfg      Config

	mu     sync.Mutex
	active map[string]bool // ticketID -> in-flight in this process
	wg     sync.WaitGroup
}

// New builds a scheduler. bus may be nil (events are then only persisted).
func New(db *sqlite.DB, policies *policy.Set, registry *actor.Registry, rt runtime.Runtime, bus *eventbus.Hub, cfg Config) *Scheduler {
	if cfg.MaxConcurrency <= 0 {
		cfg.MaxConcurrency = 4
	}
	if cfg.TickInterval <= 0 {
		cfg.TickInterval = time.Second
	}
	return &Scheduler{db: db, policies: policies, registry: registry, rt: rt, bus: bus, cfg: cfg, active: map[string]bool{}}
}

// Run drives ticks until ctx is cancelled, then waits for supervisors to settle.
func (s *Scheduler) Run(ctx context.Context) error {
	t := time.NewTicker(s.cfg.TickInterval)
	defer t.Stop()
	for {
		if _, err := s.Tick(ctx); err != nil {
			s.publish(eventbus.Event{Type: "scheduler.error", Message: err.Error()})
		}
		select {
		case <-ctx.Done():
			s.wg.Wait()
			return ctx.Err()
		case <-t.C:
		}
	}
}

// Wait blocks until all in-flight supervisors finish (used by tests and shutdown).
func (s *Scheduler) Wait() { s.wg.Wait() }

// Tick performs one scheduling pass: claim and dispatch eligible nodes up to the
// concurrency limit. It returns the number of runs started.
func (s *Scheduler) Tick(ctx context.Context) (int, error) {
	// Eligible nodes in value order (ADR 0039), not FIFO-by-id. The ordering is
	// the store's shared surface, so the scheduler and the human CLI reads agree
	// on "ready, in priority order" (ADR 0041).
	eligible, err := s.db.ListEligibleOrdered()
	if err != nil {
		return 0, err
	}
	started := 0
	for _, tk := range eligible {
		if !s.capacityAvailable() {
			break // cheap early-out; reserve() is the authoritative gate
		}
		_, ok, err := s.dispatch(ctx, tk)
		if err != nil {
			return started, err
		}
		if ok {
			started++
		}
	}
	return started, nil
}

// ErrNotDispatched explains why a specific node could not be dispatched.
var ErrNotDispatched = errors.New("node could not be dispatched")

// RunOnce attempts to claim and dispatch a single node now (gw run once). It
// returns the started run, or an error if the node is not eligible, no actor is
// authorized, or capacity is full.
func (s *Scheduler) RunOnce(ctx context.Context, ticketID string) (*sqlite.Run, error) {
	tk, err := s.db.GetTicket(ticketID)
	if err != nil {
		return nil, err
	}
	run, ok, err := s.dispatch(ctx, tk)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrNotDispatched
	}
	return run, nil
}

// dispatch selects an authorized actor, atomically reserves a concurrency slot,
// claims the node transactionally, and launches a supervisor. It reports whether
// a run was started. The slot is reserved before StartRun and released on
// failure, so concurrent dispatchers (the loop and HTTP run-triggers) can never
// exceed MaxConcurrency (ADR 0026).
func (s *Scheduler) dispatch(ctx context.Context, tk *ticket.Ticket) (*sqlite.Run, bool, error) {
	a, mode, actionType := s.selectActor(tk)
	if a == nil {
		return nil, false, nil // no authorized actor for this node
	}
	if !s.reserve(tk.ID) {
		return nil, false, nil // at capacity or already in-flight
	}
	snapshot, _ := json.Marshal(a)
	r, _, err := s.db.StartRun(sqlite.StartRunParams{
		TicketID: tk.ID, ActorID: a.ID, ActorSnapshot: string(snapshot),
		Mode: mode, Runtime: s.rt.Name(), Model: s.cfg.Model, TTL: s.cfg.LeaseTTL,
	})
	if err != nil {
		s.markDone(tk.ID) // release the reservation
		// Lost the claim race or no longer eligible: not an error to the caller.
		if errors.Is(err, sqlite.ErrNotEligible) || errors.Is(err, sqlite.ErrAlreadyLeased) {
			return nil, false, nil
		}
		return nil, false, err
	}
	s.publish(eventbus.Event{Type: "run.started", RunID: r.ID, TicketID: tk.ID,
		Payload: map[string]any{"actor_id": a.ID, "mode": string(mode), "action": actionType}})
	s.wg.Add(1)
	go s.supervise(ctx, r, tk.ID, a.ID)
	return r, true, nil
}

// selectActor picks an authorized AI actor for a node, returning the actor, the
// run mode, and the gated action type. A requested actor is a hint resolved to a
// concrete instance; every candidate must both hold the work-type capability and
// pass the policy claim gate (ADR 0023/0028/0029). Humans are not auto-dispatched.
func (s *Scheduler) selectActor(tk *ticket.Ticket) (*actor.Actor, run.Mode, string) {
	mode := run.ModeImplementation
	actionType := "execute"
	if tk.NodeType == ticket.NodeComposite {
		mode = run.ModePlanning
		actionType = "decompose"
	}

	var candidates []*actor.Actor
	if tk.RequestedActor != "" {
		if a, ok := s.registry.Resolve(tk.RequestedActor); ok {
			candidates = []*actor.Actor{a}
		}
	} else {
		for i := range s.registry.Actors {
			candidates = append(candidates, &s.registry.Actors[i])
		}
	}

	for _, a := range candidates {
		if a.Type != actor.TypeAIAgent {
			continue // scheduler dispatches to AI runtimes; humans claim via CLI
		}
		if !a.CanClaim(tk.WorkType) {
			continue
		}
		action := policy.Action{Type: actionType, Actor: a, WorkType: tk.WorkType, Scope: risk.Scope{}}
		if s.policies.AuthorizeClaim(action).Outcome == policy.OutcomeAllow {
			return a, mode, actionType
		}
	}
	return nil, "", ""
}

// supervise drives one run to completion: it streams runtime events to the store
// and bus, renews the lease via heartbeat, and on success completes the run,
// releases the lease, and moves the node to review (prepared work / proposal
// awaiting its gate). ctx cancellation interrupts the run.
func (s *Scheduler) supervise(ctx context.Context, r *sqlite.Run, ticketID, actorID string) {
	defer s.wg.Done()
	defer s.markDone(ticketID)

	hbCtx, stopHB := context.WithCancel(ctx)
	go s.heartbeat(hbCtx, r.ID, ticketID)
	defer stopHB()

	sink := func(ev runtime.Event) {
		_, _ = s.db.AppendRunEvent(r.ID, ev.Type, ev.Message, ev.Payload)
		// Mirror to the per-run events.ndjson (ADR 0027); a log failure is observable
		// but never fails the run.
		if err := appendRunEventLog(s.cfg.RunLogDir, r.ID, ev); err != nil {
			s.publish(eventbus.Event{Type: "run.error", RunID: r.ID, TicketID: ticketID, Message: "events.ndjson: " + err.Error()})
		}
		s.publish(eventbus.Event{Type: "run." + ev.Type, RunID: r.ID, TicketID: ticketID, Message: ev.Message})
	}
	spec := runtime.Spec{RunID: r.ID, TicketID: ticketID, Mode: r.Mode, ActorID: actorID,
		Runtime: r.Runtime, Model: r.Model, Workspace: r.WorkspacePath}

	res, err := s.rt.Run(ctx, spec, sink)
	stopHB()

	// Cleanup transitions are best-effort but observable: a failure here (e.g. a
	// concurrent CancelRun already moved the run/node — the stub's window is
	// microseconds, but real cancel in Phase 4 must signal the running agent)
	// publishes a run.error rather than silently stranding the node. Recovery and
	// lease TTL are the backstop.
	if err != nil {
		s.cleanup("interrupt run", r.ID, ticketID, s.db.SetRunStatus(r.ID, run.StatusInterrupted, actorID))
		s.cleanup("release lease", r.ID, ticketID, s.db.ReleaseLease(ticketID, r.ID))
		s.publish(eventbus.Event{Type: "run.interrupted", RunID: r.ID, TicketID: ticketID, Message: err.Error()})
		return
	}

	s.cleanup("complete run", r.ID, ticketID, s.db.SetRunStatus(r.ID, run.StatusCompleted, actorID))
	s.cleanup("release lease", r.ID, ticketID, s.db.ReleaseLease(ticketID, r.ID))
	// Prepared work (leaf) or a decomposition proposal (composite) awaits review.
	s.cleanup("transition to review", r.ID, ticketID, s.db.TransitionTicket(ticketID, ticket.StatusReview, actorID))
	s.publish(eventbus.Event{Type: "run.completed", RunID: r.ID, TicketID: ticketID, Message: res.LastMessage})
}

// cleanup publishes a run.error event when a best-effort post-run state change
// fails, so the inconsistency is observable instead of silent.
func (s *Scheduler) cleanup(op, runID, ticketID string, err error) {
	if err != nil {
		s.publish(eventbus.Event{Type: "run.error", RunID: runID, TicketID: ticketID, Message: op + ": " + err.Error()})
	}
}

// heartbeat renews the run's lease on the configured interval until cancelled.
func (s *Scheduler) heartbeat(ctx context.Context, runID, ticketID string) {
	if s.cfg.Heartbeat <= 0 {
		return
	}
	t := time.NewTicker(s.cfg.Heartbeat)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			_, _ = s.db.RenewLease(ticketID, runID, s.cfg.LeaseTTL)
		}
	}
}

func (s *Scheduler) publish(ev eventbus.Event) {
	if s.bus != nil {
		s.bus.Publish(ev)
	}
}

func (s *Scheduler) capacityAvailable() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.active) < s.cfg.MaxConcurrency
}

// reserve atomically claims a concurrency slot for ticketID, returning false if
// the node is already in-flight or the pool is at capacity. This check-and-set
// under one lock is what keeps concurrent dispatchers within MaxConcurrency.
func (s *Scheduler) reserve(ticketID string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.active[ticketID] || len(s.active) >= s.cfg.MaxConcurrency {
		return false
	}
	s.active[ticketID] = true
	return true
}

func (s *Scheduler) markDone(ticketID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.active, ticketID)
}

// ActiveCount reports the number of in-flight runs (for tests/diagnostics).
func (s *Scheduler) ActiveCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.active)
}
