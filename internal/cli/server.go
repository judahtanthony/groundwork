package cli

import (
	"context"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"groundwork/internal/actor"
	"groundwork/internal/eventbus"
	"groundwork/internal/policy"
	"groundwork/internal/runtime"
	"groundwork/internal/scheduler"
	"groundwork/internal/server"
)

func newServerCmd() *Command {
	return &Command{
		Name:  "server",
		Usage: "Run the localhost coordinator (HTTP API + SSE)",
		Run:   runServer,
	}
}

// runServer starts the coordinator: the HTTP server plus the scheduler loop. It
// blocks until interrupted. The bind address comes from config unless --addr
// overrides it.
func runServer(ctx *Context, args []string) error {
	fs := ctx.NewFlagSet("gw server")
	var addr string
	var noScheduler bool
	fs.StringVar(&addr, "addr", "", "bind address override (default: config server.addr)")
	fs.BoolVar(&noScheduler, "no-scheduler", false,
		"run the coordinator without the scheduler, so eligible nodes are not auto-claimed "+
			"(human-performed work owns the lifecycle; ADR 0033)")
	if _, err := parseFlags(fs, args); err != nil {
		return err
	}

	p, db, err := ctx.openStore()
	if err != nil {
		return err
	}
	defer db.Close()

	if addr == "" {
		addr = p.Config.Server.Addr
	}

	// Cold start: rebuild nodes from committed exports when the store is empty
	// (recovery.md). Then reconcile any runs/leases left by a previous process.
	if has, err := db.HasTickets(); err == nil && !has {
		if n, err := importExports(db, p.TicketsDir()); err == nil && n > 0 {
			ctx.Stderr.Write([]byte("gw: imported " + strconv.Itoa(n) + " ticket(s) from exports\n"))
		}
	}
	if rep, err := db.ReconcileStartup(); err != nil {
		return &Error{Code: "recovery_error", Message: err.Error()}
	} else if rep.InterruptedRuns > 0 || rep.ReleasedLeases > 0 {
		ctx.Stderr.Write([]byte("gw: recovery interrupted " + strconv.Itoa(rep.InterruptedRuns) +
			" run(s), released " + strconv.Itoa(rep.ReleasedLeases) + " lease(s)\n"))
	}

	// Load policy and the actor registry for scheduling decisions. Missing or
	// invalid files surface as warnings/errors rather than silently disabling
	// gates.
	policies, pwarn, err := policy.Load(p.PoliciesDir())
	if err != nil {
		return &Error{Code: "policy_error", Message: err.Error()}
	}
	registry, awarn, err := actor.Load(p.ActorsPath())
	if err != nil {
		return &Error{Code: "actors_error", Message: err.Error()}
	}
	for _, w := range append(pwarn, awarn...) {
		ctx.Stderr.Write([]byte("gw: warning: " + w + "\n"))
	}

	bus := eventbus.New(0)
	defer bus.Close()

	srv := server.New(db, p, Version)
	srv.SetBus(bus)
	srv.SetApprovals(server.NewApprovalService(db, policies, registry))

	sigCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// The scheduler auto-claims eligible nodes and dispatches them to AI actors
	// through the runtime. In M3, human-performed work owns the node lifecycle via
	// manual transitions (ADR 0033); --no-scheduler runs the coordinator (API,
	// gates, approvals, landing) without auto-dispatch so the human is not raced.
	if noScheduler {
		ctx.Stderr.Write([]byte("gw: scheduler disabled (--no-scheduler); nodes are not auto-claimed\n"))
	} else {
		sched := scheduler.New(db, policies, registry, runtime.Stub{}, bus, scheduler.Config{
			MaxConcurrency: p.Config.MaxConcurrency,
			LeaseTTL:       p.Config.Lease.TTL.Duration(),
			Heartbeat:      p.Config.Lease.Heartbeat.Duration(),
			TickInterval:   time.Second,
		})
		srv.SetScheduler(sched)
		go func() { _ = sched.Run(sigCtx) }()
	}

	if err := srv.Serve(sigCtx, addr, ctx.Stderr); err != nil {
		return &Error{Code: "server_error", Message: err.Error()}
	}
	return nil
}
