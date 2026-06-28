package runtime

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
)

// WorkspaceProvider provisions and removes the isolated tree a run executes in
// (ADR 0059). It is satisfied by internal/worktree.Manager (adapted at wiring
// time); the runtime depends only on this seam so it stays testable.
type WorkspaceProvider interface {
	// Provision creates the run's isolated worktree from base and returns its path.
	Provision(runID, base string) (string, error)
}

// BaseResolver returns the integration base commit/ref a run's worktree is cut
// from (the root integration branch tip, else the default branch). It is called
// at dispatch so the base is fixed for the run (ADR 0059).
type BaseResolver func(spec Spec) string

// execLauncher runs the configured agent command in the run's worktree and
// streams its output as events (T-0502). It validates that the working directory
// is the provisioned, contained worktree before launching, so a misconfigured run
// can never execute the agent in the repo root or an arbitrary directory.
func execLauncher(ctx context.Context, spec Spec, sink Sink, cfg Config) (Result, error) {
	dir, err := validateWorkspace(spec.Workspace, cfg.WorktreeRoot)
	if err != nil {
		return Result{Status: "error"}, err
	}
	emit(sink, Event{Type: "claimed", Message: "codex claimed " + spec.TicketID,
		Payload: map[string]any{"model": spec.Model, "sandbox": cfg.Sandbox, "workspace": dir}})

	cmd := exec.CommandContext(ctx, cfg.Command, codexArgs(cfg, spec)...)
	cmd.Dir = dir
	cmd.Env = os.Environ()

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return Result{Status: "error"}, err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return Result{Status: "error"}, err
	}
	if err := cmd.Start(); err != nil {
		return Result{Status: "error"}, fmt.Errorf("codex: launch %q: %w", cfg.Command, err)
	}
	emit(sink, Event{Type: "working", Message: "codex running"})

	var last string
	var mu sync.Mutex
	var wg sync.WaitGroup
	stream := func(r io.Reader, typ string) {
		defer wg.Done()
		sc := bufio.NewScanner(r)
		sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
		for sc.Scan() {
			line := sc.Text()
			mu.Lock()
			if strings.TrimSpace(line) != "" {
				last = line
			}
			mu.Unlock()
			emit(sink, Event{Type: typ, Message: line})
		}
	}
	wg.Add(2)
	go stream(stdout, "output")
	go stream(stderr, "stderr")
	wg.Wait()

	if err := cmd.Wait(); err != nil {
		// A cancelled context is an interrupt, not an agent failure.
		if ctx.Err() != nil {
			return Result{Status: "interrupted", LastMessage: last}, ctx.Err()
		}
		emit(sink, Event{Type: "failed", Message: err.Error()})
		return Result{Status: "failed", LastMessage: last}, fmt.Errorf("codex run failed: %w", err)
	}
	emit(sink, Event{Type: "produced", Message: "codex completed"})
	emit(sink, Event{Type: "awaiting_gate", Message: "awaiting approval gate"})
	return Result{Status: "produced", LastMessage: last}, nil
}

// validateWorkspace ensures dir is a non-empty, existing directory and, when a
// worktree root is configured, that it is contained within it (ADR 0059).
func validateWorkspace(dir, root string) (string, error) {
	if strings.TrimSpace(dir) == "" {
		return "", fmt.Errorf("codex: no workspace; a run worktree is required")
	}
	abs, err := filepath.Abs(dir)
	if err != nil {
		return "", err
	}
	fi, err := os.Stat(abs)
	if err != nil || !fi.IsDir() {
		return "", fmt.Errorf("codex: workspace %q is not a directory", dir)
	}
	if root != "" {
		rootAbs, err := filepath.Abs(root)
		if err != nil {
			return "", err
		}
		if abs != rootAbs && !strings.HasPrefix(abs, rootAbs+string(filepath.Separator)) {
			return "", fmt.Errorf("codex: workspace %q escapes worktree root %q", abs, rootAbs)
		}
	}
	return abs, nil
}

// codexArgs builds the agent command arguments: configured args plus the model
// when one was resolved. The exact agent CLI contract is intentionally thin in
// v1 — the prompt/turn protocol is supplied by config (cfg.Args).
func codexArgs(cfg Config, spec Spec) []string {
	args := append([]string(nil), cfg.Args...)
	if spec.Model != "" {
		args = append(args, "--model", spec.Model)
	}
	return args
}

func emit(sink Sink, ev Event) {
	if sink != nil {
		sink(ev)
	}
}
