package cli

import (
	"errors"
	"fmt"

	"groundwork/internal/store/sqlite"
)

// newRunCmd builds the `gw run` subtree. Every subcommand drives live run
// control and therefore requires a running coordinator (ADR 0031).
func newRunCmd() *Command {
	return &Command{
		Name:  "run",
		Usage: "Inspect and control runs (requires the coordinator)",
		Sub: []*Command{
			{Name: "once", Usage: "Dispatch one node now", Args: "<ticket-id>", Run: runRunOnce},
			{Name: "next", Usage: "Dispatch the next eligible node(s)", Run: runRunNext},
			{Name: "list", Usage: "List runs", Run: runRunList},
			{Name: "show", Usage: "Show a run and its events", Args: "<run-id>", Run: runRunShow},
			{Name: "pause", Usage: "Pause a run", Args: "<run-id>", Run: runRunPause},
			{Name: "resume", Usage: "Resume a run", Args: "<run-id>", Run: runRunResume},
			{Name: "cancel", Usage: "Cancel a run", Args: "<run-id>", Run: runRunCancel},
		},
	}
}

func runRunOnce(ctx *Context, args []string) error {
	pos, err := positional(ctx, "gw run once", args, 1, "usage: gw run once <ticket-id>")
	if err != nil {
		return err
	}
	c, err := ctx.requireCoordinator()
	if err != nil {
		return err
	}
	run, err := c.RunOnce(pos[0])
	if err != nil {
		return runError(err, pos[0])
	}
	if ctx.JSON {
		return ctx.PrintJSON(run)
	}
	fmt.Fprintf(ctx.Stdout, "Started %s for %s (%s)\n", run.ID, run.TicketID, run.Status)
	return nil
}

func runRunNext(ctx *Context, args []string) error {
	if _, err := positional(ctx, "gw run next", args, 0, ""); err != nil {
		return err
	}
	c, err := ctx.requireCoordinator()
	if err != nil {
		return err
	}
	n, err := c.RunNext()
	if err != nil {
		return &Error{Code: "run_failed", Message: err.Error()}
	}
	if ctx.JSON {
		return ctx.PrintJSON(map[string]int{"started": n})
	}
	fmt.Fprintf(ctx.Stdout, "Started %d run(s)\n", n)
	return nil
}

func runRunList(ctx *Context, args []string) error {
	if _, err := positional(ctx, "gw run list", args, 0, ""); err != nil {
		return err
	}
	c, err := ctx.requireCoordinator()
	if err != nil {
		return err
	}
	runs, err := c.ListRuns()
	if err != nil {
		return &Error{Code: "list_failed", Message: err.Error()}
	}
	if ctx.JSON {
		return ctx.PrintJSON(runs)
	}
	if len(runs) == 0 {
		fmt.Fprintln(ctx.Stdout, "No runs.")
		return nil
	}
	fmt.Fprintf(ctx.Stdout, "%-8s  %-8s  %-12s  %-12s  %s\n", "RUN", "TICKET", "MODE", "STATUS", "ACTOR")
	for _, r := range runs {
		fmt.Fprintf(ctx.Stdout, "%-8s  %-8s  %-12s  %-12s  %s\n", r.ID, r.TicketID, r.Mode, r.Status, r.ActorID)
	}
	return nil
}

func runRunShow(ctx *Context, args []string) error {
	pos, err := positional(ctx, "gw run show", args, 1, "usage: gw run show <run-id>")
	if err != nil {
		return err
	}
	c, err := ctx.requireCoordinator()
	if err != nil {
		return err
	}
	run, err := c.GetRun(pos[0])
	if err != nil {
		return runError(err, pos[0])
	}
	events, err := c.RunEvents(pos[0])
	if err != nil {
		return &Error{Code: "list_failed", Message: err.Error()}
	}
	if ctx.JSON {
		return ctx.PrintJSON(struct {
			*sqlite.Run
			Events []sqlite.RunEvent `json:"events"`
		}{run, events})
	}
	w := ctx.Stdout
	fmt.Fprintf(w, "%s  ticket=%s  mode=%s  status=%s  actor=%s\n", run.ID, run.TicketID, run.Mode, run.Status, run.ActorID)
	fmt.Fprintf(w, "  started:  %s\n", run.StartedAt)
	if run.CompletedAt != "" {
		fmt.Fprintf(w, "  completed: %s\n", run.CompletedAt)
	}
	fmt.Fprintln(w, "  events:")
	for _, e := range events {
		fmt.Fprintf(w, "    %s  %s\n", e.EventType, e.CreatedAt)
	}
	return nil
}

func runRunPause(ctx *Context, args []string) error {
	return runControl(ctx, "gw run pause", args, "pause")
}
func runRunResume(ctx *Context, args []string) error {
	return runControl(ctx, "gw run resume", args, "resume")
}
func runRunCancel(ctx *Context, args []string) error {
	return runControl(ctx, "gw run cancel", args, "cancel")
}

func runControl(ctx *Context, usage string, args []string, op string) error {
	pos, err := positional(ctx, usage, args, 1, "usage: "+usage+" <run-id>")
	if err != nil {
		return err
	}
	c, err := ctx.requireCoordinator()
	if err != nil {
		return err
	}
	var run *sqlite.Run
	switch op {
	case "pause":
		run, err = c.PauseRun(pos[0])
	case "resume":
		run, err = c.ResumeRun(pos[0])
	case "cancel":
		run, err = c.CancelRun(pos[0])
	}
	if err != nil {
		return runError(err, pos[0])
	}
	if ctx.JSON {
		return ctx.PrintJSON(run)
	}
	fmt.Fprintf(ctx.Stdout, "%s -> %s\n", run.ID, run.Status)
	return nil
}

// positional parses flags and enforces an exact positional-argument count.
func positional(ctx *Context, name string, args []string, want int, usage string) ([]string, error) {
	fs := ctx.NewFlagSet(name)
	pos, err := parseFlags(fs, args)
	if err != nil {
		return nil, err
	}
	if len(pos) < want {
		return nil, &Error{Code: "invalid_args", Message: usage}
	}
	return pos, nil
}

// runError maps a coordinator/store error to a CLI error, distinguishing not-found.
func runError(err error, id string) error {
	if errors.Is(err, sqlite.ErrNotFound) {
		return &Error{Code: "not_found", Message: fmt.Sprintf("run %q not found", id)}
	}
	return &Error{Code: "run_failed", Message: err.Error()}
}
