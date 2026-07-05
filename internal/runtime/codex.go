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
	Sandbox      string // sandbox mode (read-only|workspace-write|danger-full-access), default workspace-write
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
	base      BaseResolver      // resolves the provision base (resume checkpoint, else "")
	diffBase  BaseResolver      // resolves the integration base the diff is measured against
}

// NewCodex builds the Codex adapter with cfg, applying defaults. Until the
// process launcher is installed (WithLauncher / T-0502) it uses a records-only
// launch so dispatch stays functional.
func NewCodex(cfg Config) *Codex {
	if cfg.Command == "" {
		cfg.Command = "codex"
	}
	// Default to autonomous execution bounded to the run's worktree. `codex exec`
	// is inherently non-interactive (it never prompts), so the sandbox is the only
	// control: workspace-write confines writes to the run's worktree. Groundwork
	// adds the outer boundary (worktree isolation + envelope enforcement at
	// landing), so the agent runs unattended within scope.
	if cfg.Sandbox == "" {
		cfg.Sandbox = "workspace-write"
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

// WithDiffBase sets the resolver for the integration base the run's diff is
// measured against (ADR 0059). This must be the node's integration target, NOT
// the provision base: a resumed run is provisioned from a prior checkpoint but
// its diff must still cover everything since the integration base, or files an
// interrupted run changed would escape envelope enforcement (review finding #3).
// When unset, the diff is measured against the provision base.
func (c *Codex) WithDiffBase(r BaseResolver) *Codex {
	cp := *c
	cp.diffBase = r
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
	provBase := ""    // where the worktree is cut from (resume checkpoint, else "")
	diffBase := ""    // the integration base the run's net diff is measured against
	if c.workspace != nil {
		if c.base != nil {
			provBase = c.base(spec)
		}
		path, err := c.workspace.Provision(spec.RunID, provBase)
		if err != nil {
			return Result{Status: "error"}, fmt.Errorf("codex: provision worktree: %w", err)
		}
		spec.Workspace = path
		// The diff is measured against the integration base, NOT the provision base,
		// so a resumed run still reports everything since the integration target.
		diffBase = provBase
		if c.diffBase != nil {
			diffBase = c.diffBase(spec)
		}
		emit(sink, Event{Type: "worktree", Message: "provisioned " + path,
			Payload: map[string]any{"branch": "gw/run/" + spec.RunID, "base": provBase, "diff_base": diffBase}})
	}

	res, err := c.launch(ctx, spec, sink, c.cfg)
	if err != nil {
		return res, err
	}

	// Capture the run's changed-file set from its worktree as the authoritative
	// diff for gate inputs (ADR 0059) and as run evidence.
	if c.workspace != nil {
		files, diff, derr := c.workspace.Diff(spec.RunID, diffBase)
		if derr != nil {
			emit(sink, Event{Type: "run.error", Message: "capture diff: " + derr.Error()})
		} else {
			res.ChangedFiles = files
			res.Diff = diff
			emit(sink, Event{Type: "diff", Message: "captured changed files",
				Payload: map[string]any{"changed_files": len(files)}})
		}
		// Commit the work as a checkpoint on the run branch so it is durable for
		// landing (squash) and resume (ADR 0015). Diff already staged the worktree.
		if sha, cerr := c.workspace.Checkpoint(spec.RunID, "checkpoint "+spec.TicketID); cerr != nil {
			emit(sink, Event{Type: "run.error", Message: "checkpoint: " + cerr.Error()})
		} else if sha != "" {
			emit(sink, Event{Type: "checkpoint", Message: sha,
				Payload: map[string]any{"branch": "gw/run/" + spec.RunID}})
		}
	}
	return res, nil
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
