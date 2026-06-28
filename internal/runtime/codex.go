package runtime

import (
	"context"
	"fmt"
)

// Config configures the Codex adapter (ADR 0027). Command/Model/Sandbox come
// from project config and the per-run actor snapshot; the coordinator fills the
// per-attempt Spec (worktree, actor, model) at dispatch.
type Config struct {
	Command      string // codex executable, default "codex"
	Model        string // default model when a Spec carries none
	Sandbox      string // sandbox mode passed to the agent (e.g. workspace-write)
	Args         []string
	WorktreeRoot string // when set, the launcher requires the workspace to be contained here
}

// LaunchFunc runs one agent attempt for the Codex adapter, emitting events to
// sink. It is the seam the real process launcher fills (T-0502); the shell ships
// a records-only default so the coordinator loop runs end-to-end and config
// selection is exercised before the process integration lands.
type LaunchFunc func(ctx context.Context, spec Spec, sink Sink, cfg Config) (Result, error)

// Codex is the real runtime adapter (ADR 0027), selectable by config. The shell
// (T-0501) wires selection, config, and actor-configured launch; the isolated
// worktree (T-0506) and process exec (T-0502) plug into the LaunchFunc seam.
type Codex struct {
	cfg       Config
	launch    LaunchFunc
	workspace WorkspaceProvider // optional; provisions the run's isolated worktree
	base      BaseResolver      // resolves the integration base for the worktree
}

// NewCodex builds the Codex adapter with cfg, applying defaults. Until the
// process launcher is installed (WithLauncher / T-0502) it uses a records-only
// launch so dispatch stays functional.
func NewCodex(cfg Config) *Codex {
	if cfg.Command == "" {
		cfg.Command = "codex"
	}
	return &Codex{cfg: cfg, launch: recordsOnlyLaunch}
}

// WithLauncher returns a copy of the adapter using launch (the real process
// launcher, T-0502, slots in here). A nil launch is ignored.
func (c *Codex) WithLauncher(launch LaunchFunc) *Codex {
	if launch == nil {
		return c
	}
	cp := *c
	cp.launch = launch
	return &cp
}

// WithExec returns a copy of the adapter that launches the real agent process
// (T-0502) instead of the records-only shell.
func (c *Codex) WithExec() *Codex { return c.WithLauncher(execLauncher) }

// WithWorkspace returns a copy that provisions an isolated worktree per run from
// the resolved integration base (ADR 0059), setting the run's workspace before
// launch. base may be nil (the provider then cuts from HEAD).
func (c *Codex) WithWorkspace(p WorkspaceProvider, base BaseResolver) *Codex {
	cp := *c
	cp.workspace = p
	cp.base = base
	return &cp
}

// Name identifies the codex runtime.
func (c *Codex) Name() string { return "codex" }

// Run executes one attempt. It resolves the effective model from the Spec (the
// coordinator's actor selection) falling back to config, then delegates to the
// configured launcher.
func (c *Codex) Run(ctx context.Context, spec Spec, sink Sink) (Result, error) {
	if spec.Model == "" {
		spec.Model = c.cfg.Model
	}
	if c.launch == nil {
		return Result{Status: "error"}, fmt.Errorf("codex: no launcher configured")
	}
	// Provision an isolated worktree for the run when a provider is configured, so
	// the agent executes against a private tree from a fixed base (ADR 0059). The
	// worktree (and its gw/run/<id> branch) persists past the run for diff capture
	// and landing; abandoned worktrees are reclaimed by recovery reconciliation.
	if c.workspace != nil {
		base := ""
		if c.base != nil {
			base = c.base(spec)
		}
		path, err := c.workspace.Provision(spec.RunID, base)
		if err != nil {
			return Result{Status: "error"}, fmt.Errorf("codex: provision worktree: %w", err)
		}
		spec.Workspace = path
		emit(sink, Event{Type: "worktree", Message: "provisioned " + path,
			Payload: map[string]any{"branch": "gw/run/" + spec.RunID, "base": base}})
	}
	return c.launch(ctx, spec, sink, c.cfg)
}

// recordsOnlyLaunch is the shell's default launcher: it emits the same synthetic
// lifecycle as the stub and writes no code, so the scheduler → run → events →
// gate → landing loop runs before the real process integration (T-0502) lands.
func recordsOnlyLaunch(ctx context.Context, spec Spec, sink Sink, cfg Config) (Result, error) {
	events := []Event{
		{Type: "claimed", Message: "codex claimed " + spec.TicketID, Payload: map[string]any{"model": spec.Model, "sandbox": cfg.Sandbox}},
		{Type: "working", Message: "codex preparing (records-only shell; launch is T-0502)"},
		{Type: "produced", Message: "produced records (no code)"},
		{Type: "awaiting_gate", Message: "awaiting approval gate"},
	}
	for _, ev := range events {
		if err := ctx.Err(); err != nil {
			return Result{Status: "interrupted"}, err
		}
		if sink != nil {
			sink(ev)
		}
	}
	return Result{Status: "produced", LastMessage: "produced records (no code)"}, nil
}
