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
	"groundwork/internal/git"
	"groundwork/internal/policy"
	"groundwork/internal/runtime"
	"groundwork/internal/scheduler"
	"groundwork/internal/server"
	"groundwork/internal/worktree"
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
	fs.StringVar(&addr, "addr", "", "bind address override (default: config server.addr)")
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
	hadTickets, _ := db.HasTickets()
	if !hadTickets {
		if n, err := importExports(db, p.TicketsDir()); err == nil && n > 0 {
			ctx.Stderr.Write([]byte("gw: imported " + strconv.Itoa(n) + " ticket(s) from exports\n"))
		}
	}

	// Enable filesystem write-through now that the store reflects files (ADR 0053):
	// from here, durable mutations rewrite their sidecars before reporting success.
	db.SetExportDir(p.TicketsDir())

	// A store that SURVIVED a restart may hold durable mutations that never reached
	// files (a crash between SQLite commit and the sidecar write). Surface that as
	// recovery_needed rather than silently trusting SQLite (ADR 0053). A freshly
	// rebuilt store matches its files by construction, so skip the check then.
	if hadTickets {
		if drep, err := db.DetectFileDivergence(); err != nil {
			return &Error{Code: "recovery_error", Message: err.Error()}
		} else if len(drep.Diverged) > 0 {
			ctx.Stderr.Write([]byte("gw: warning: " + strconv.Itoa(len(drep.Diverged)) +
				" ticket(s) diverge from their sidecars (unexported durable mutation); flagged recovery_needed — rebuild from files to repair\n"))
		}
	}

	if rep, err := db.ReconcileStartup(); err != nil {
		return &Error{Code: "recovery_error", Message: err.Error()}
	} else if rep.InterruptedRuns > 0 || rep.ReleasedLeases > 0 {
		ctx.Stderr.Write([]byte("gw: recovery interrupted " + strconv.Itoa(rep.InterruptedRuns) +
			" run(s), released " + strconv.Itoa(rep.ReleasedLeases) + " lease(s)\n"))
	}

	// Rebuild live approval/decision queues from durable ticket records, and
	// surface recovery_needed for any stranded blocked/review/rework ticket
	// (ADR 0051). Safe to run every boot; idempotent.
	if qrep, err := db.RebuildDurableQueues(); err != nil {
		return &Error{Code: "recovery_error", Message: err.Error()}
	} else if qrep.ApprovalsRecreated > 0 || qrep.RecoveryNeeded > 0 {
		ctx.Stderr.Write([]byte("gw: recovery recreated " + strconv.Itoa(qrep.ApprovalsRecreated) +
			" approval(s), flagged " + strconv.Itoa(qrep.RecoveryNeeded) + " recovery_needed\n"))
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

	// Select the runtime adapter from config (ADR 0027): records-only stub or the
	// Codex adapter. The Codex adapter runs in an isolated worktree per run.
	rt, err := runtime.Select(p.Config.Runtime, runtime.Config{
		Model:        p.Config.Model,
		Sandbox:      p.Config.Sandbox,
		WorktreeRoot: p.WorktreesDir(),
	})
	if err != nil {
		return &Error{Code: "runtime_error", Message: err.Error()}
	}
	// The Codex adapter executes in an isolated git worktree per run (ADR 0059).
	// When the project is a git work tree, give it the real process launcher and a
	// worktree provider; otherwise it stays the records-only shell.
	if codex, ok := rt.(*runtime.Codex); ok {
		if repo, gerr := git.Open(p.Root); gerr == nil {
			mgr := worktree.NewManager(repo, p.WorktreesDir())
			rt = codex.WithExec().WithWorkspace(worktreeProvider{mgr}, nil)
		} else {
			ctx.Stderr.Write([]byte("gw: warning: project is not a git work tree; codex runs records-only\n"))
		}
	}
	ctx.Stderr.Write([]byte("gw: runtime " + rt.Name() + "\n"))
	sched := scheduler.New(db, policies, registry, rt, bus, scheduler.Config{
		MaxConcurrency: p.Config.MaxConcurrency,
		LeaseTTL:       p.Config.Lease.TTL.Duration(),
		Heartbeat:      p.Config.Lease.Heartbeat.Duration(),
		TickInterval:   time.Second,
		Model:          p.Config.Model,
		RunLogDir:      p.RunsDir(),
	})

	srv := server.New(db, p, Version)
	srv.SetScheduler(sched)
	srv.SetBus(bus)
	srv.SetApprovals(server.NewApprovalService(db, policies, registry))

	sigCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// The scheduler only dispatches a node when the trust policy authorizes an AI
	// actor to claim it (allow_claim). In M3 the project policy authorizes no AI
	// claims, so the scheduler finds no actor and human-performed work owns the
	// lifecycle (ADR 0033); loosening allow_claim is what makes work available.
	go func() { _ = sched.Run(sigCtx) }()

	if err := srv.Serve(sigCtx, addr, ctx.Stderr); err != nil {
		return &Error{Code: "server_error", Message: err.Error()}
	}
	return nil
}

// worktreeProvider adapts worktree.Manager to runtime.WorkspaceProvider so the
// Codex adapter can provision an isolated worktree per run (ADR 0059).
type worktreeProvider struct{ m *worktree.Manager }

func (w worktreeProvider) Provision(runID, base string) (string, error) {
	p, err := w.m.Provision(runID, base)
	if err != nil {
		return "", err
	}
	return p.Path, nil
}

func (w worktreeProvider) Diff(runID, base string) ([]string, string, error) {
	return w.m.Diff(runID, base)
}
