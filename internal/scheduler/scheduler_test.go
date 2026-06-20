package scheduler

import (
	"context"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"groundwork/internal/actor"
	"groundwork/internal/policy"
	"groundwork/internal/run"
	"groundwork/internal/runtime"
	"groundwork/internal/store/sqlite"
	"groundwork/internal/ticket"
)

func newDB(t *testing.T) *sqlite.DB {
	t.Helper()
	db, err := sqlite.Open(filepath.Join(t.TempDir(), "state.sqlite"))
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	if err := db.Migrate(); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return db
}

func testRegistry() *actor.Registry {
	return &actor.Registry{Schema: actor.SchemaVersion, Actors: []actor.Actor{
		{ID: "human.owner", Type: actor.TypeHuman, Capabilities: actor.Capabilities{WorkTypes: []string{"*"}}},
		{ID: "ai.codex.default", Type: actor.TypeAIAgent, Capabilities: actor.Capabilities{WorkTypes: []string{"technical_implementation"}}},
	}}
}

func allowCodexPolicy() *policy.Set {
	return &policy.Set{Trust: &policy.TrustPolicy{AllowClaim: []policy.Rule{{
		ID:      "codex",
		When:    policy.Match{ActorIDs: []string{"ai.codex.default"}, WorkTypes: []string{"technical_implementation"}, RiskClassAtMost: "medium"},
		Actions: []string{"execute", "decompose"},
	}}}}
}

func testConfig() Config {
	return Config{MaxConcurrency: 4, LeaseTTL: 90 * time.Second, Heartbeat: 0, TickInterval: time.Hour}
}

func createTodo(t *testing.T, db *sqlite.DB, wt string) *ticket.Ticket {
	t.Helper()
	tk := &ticket.Ticket{Title: "node", Status: ticket.StatusTodo, WorkType: wt}
	if err := db.CreateTicket(tk, "tester"); err != nil {
		t.Fatalf("create: %v", err)
	}
	return tk
}

func TestTickDispatchesEligible(t *testing.T) {
	db := newDB(t)
	tk := createTodo(t, db, "technical_implementation")
	s := New(db, allowCodexPolicy(), testRegistry(), runtime.Stub{}, nil, testConfig())

	started, err := s.Tick(context.Background())
	if err != nil {
		t.Fatalf("Tick: %v", err)
	}
	if started != 1 {
		t.Fatalf("started = %d, want 1", started)
	}
	s.Wait()

	runs, _ := db.ListRunsForTicket(tk.ID)
	if len(runs) != 1 || runs[0].Status != string(run.StatusCompleted) {
		t.Fatalf("runs = %+v, want one completed", runs)
	}
	got, _ := db.GetTicket(tk.ID)
	if got.Status != ticket.StatusReview {
		t.Errorf("ticket status = %s, want review", got.Status)
	}
	events, _ := db.ListRunEvents(runs[0].ID)
	if len(events) < 4 {
		t.Errorf("got %d events, want >= 4", len(events))
	}
}

func TestTickDeniesUnauthorizedActor(t *testing.T) {
	db := newDB(t)
	createTodo(t, db, "technical_implementation")
	// Empty trust policy => AuthorizeClaim default-denies.
	s := New(db, &policy.Set{}, testRegistry(), runtime.Stub{}, nil, testConfig())

	started, err := s.Tick(context.Background())
	if err != nil {
		t.Fatalf("Tick: %v", err)
	}
	if started != 0 {
		t.Fatalf("started = %d, want 0 (no authorized actor)", started)
	}
}

func TestTickSkipsCapabilityMismatch(t *testing.T) {
	db := newDB(t)
	createTodo(t, db, "deployment") // codex cannot claim deployment
	s := New(db, allowCodexPolicy(), testRegistry(), runtime.Stub{}, nil, testConfig())

	started, _ := s.Tick(context.Background())
	if started != 0 {
		t.Fatalf("started = %d, want 0 (capability mismatch)", started)
	}
}

func TestTickRespectsDependencies(t *testing.T) {
	db := newDB(t)
	a := createTodo(t, db, "technical_implementation")
	b := createTodo(t, db, "technical_implementation")
	if err := db.AddDependency(b.ID, a.ID, "tester"); err != nil {
		t.Fatalf("link: %v", err)
	}
	s := New(db, allowCodexPolicy(), testRegistry(), runtime.Stub{}, nil, testConfig())

	started, _ := s.Tick(context.Background())
	if started != 1 {
		t.Fatalf("started = %d, want 1 (only a eligible)", started)
	}
	s.Wait()

	if runs, _ := db.ListRunsForTicket(b.ID); len(runs) != 0 {
		t.Errorf("b should not have run while its dependency is unmet: %+v", runs)
	}
	if runs, _ := db.ListRunsForTicket(a.ID); len(runs) != 1 {
		t.Errorf("a should have one run")
	}
}

// blockingRuntime blocks in Run until release is closed, so concurrency can be
// observed deterministically.
type blockingRuntime struct {
	release chan struct{}
	mu      sync.Mutex
	started int
}

func (b *blockingRuntime) Name() string { return "blocking" }
func (b *blockingRuntime) Run(ctx context.Context, spec runtime.Spec, sink runtime.Sink) (runtime.Result, error) {
	b.mu.Lock()
	b.started++
	b.mu.Unlock()
	select {
	case <-b.release:
	case <-ctx.Done():
		return runtime.Result{}, ctx.Err()
	}
	return runtime.Result{Status: "produced"}, nil
}

func TestTickRespectsMaxConcurrency(t *testing.T) {
	db := newDB(t)
	for i := 0; i < 5; i++ {
		createTodo(t, db, "technical_implementation")
	}
	rt := &blockingRuntime{release: make(chan struct{})}
	cfg := testConfig()
	cfg.MaxConcurrency = 2
	s := New(db, allowCodexPolicy(), testRegistry(), rt, nil, cfg)

	started, err := s.Tick(context.Background())
	if err != nil {
		t.Fatalf("Tick: %v", err)
	}
	if started != 2 {
		t.Fatalf("started = %d, want 2 (capacity)", started)
	}
	if got := s.ActiveCount(); got != 2 {
		t.Fatalf("active = %d, want 2", got)
	}
	close(rt.release)
	s.Wait()
	if rt.started != 2 {
		t.Errorf("runtime ran %d times, want 2", rt.started)
	}
}

func TestConcurrentDispatchRespectsCap(t *testing.T) {
	db := newDB(t)
	for i := 0; i < 10; i++ {
		createTodo(t, db, "technical_implementation")
	}
	rt := &blockingRuntime{release: make(chan struct{})}
	cfg := testConfig()
	cfg.MaxConcurrency = 3
	s := New(db, allowCodexPolicy(), testRegistry(), rt, nil, cfg)

	// Hammer the dispatch path from many goroutines (as the loop + HTTP triggers
	// would). The atomic reserve must keep concurrent dispatchers within the cap.
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _ = s.Tick(context.Background())
		}()
	}
	wg.Wait()

	if got := s.ActiveCount(); got != 3 {
		t.Fatalf("active = %d, want 3 (cap)", got)
	}
	close(rt.release)
	s.Wait()
	if rt.started != 3 {
		t.Errorf("runtime ran %d times, want 3", rt.started)
	}
}
