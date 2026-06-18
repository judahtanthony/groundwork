package cli

import (
	"errors"
	"fmt"
	"os"

	"groundwork/internal/actor"
	"groundwork/internal/config"
)

func newActorCmd() *Command {
	return &Command{
		Name:  "actor",
		Usage: "Inspect the local actor registry",
		Sub: []*Command{
			{Name: "list", Usage: "List actors", Run: runActorList},
			{Name: "show", Usage: "Show an actor", Args: "<id>", Run: runActorShow},
			{Name: "validate", Usage: "Validate actors.yaml", Run: runActorValidate},
		},
	}
}

// loadRegistry resolves the project and parses actors.yaml.
func (ctx *Context) loadRegistry() (*config.Project, *actor.Registry, []string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, nil, nil, &Error{Code: "cwd_failed", Message: err.Error()}
	}
	p, err := config.Open(cwd, ctx.RootFlag)
	if err != nil {
		if errors.Is(err, config.ErrProjectNotFound) {
			return nil, nil, nil, &Error{Code: "no_project", Message: err.Error()}
		}
		return nil, nil, nil, &Error{Code: "config_error", Message: err.Error()}
	}
	reg, warnings, err := actor.Load(p.ActorsPath())
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil, nil, &Error{Code: "no_actors", Message: "no actors.yaml found; run \"gw init\" to scaffold one"}
		}
		return nil, nil, nil, &Error{Code: "actors_invalid", Message: err.Error()}
	}
	return p, reg, warnings, nil
}

func runActorList(ctx *Context, args []string) error {
	fs := ctx.NewFlagSet("gw actor list")
	if _, err := parseFlags(fs, args); err != nil {
		return err
	}
	_, reg, _, err := ctx.loadRegistry()
	if err != nil {
		return err
	}
	if ctx.JSON {
		return ctx.PrintJSON(reg.Actors)
	}
	fmt.Fprintf(ctx.Stdout, "%-20s  %-10s  %s\n", "ID", "TYPE", "NAME")
	for _, a := range reg.Actors {
		fmt.Fprintf(ctx.Stdout, "%-20s  %-10s  %s\n", a.ID, a.Type, a.DisplayName)
	}
	return nil
}

func runActorShow(ctx *Context, args []string) error {
	fs := ctx.NewFlagSet("gw actor show")
	pos, err := parseFlags(fs, args)
	if err != nil {
		return err
	}
	if len(pos) < 1 {
		return &Error{Code: "invalid_args", Message: "usage: gw actor show <id>"}
	}
	_, reg, _, err := ctx.loadRegistry()
	if err != nil {
		return err
	}
	a, ok := reg.Get(pos[0])
	if !ok {
		return &Error{Code: "not_found", Message: fmt.Sprintf("actor %q not found", pos[0])}
	}
	if ctx.JSON {
		return ctx.PrintJSON(a)
	}
	w := ctx.Stdout
	fmt.Fprintf(w, "%s  %s\n", a.ID, a.DisplayName)
	fmt.Fprintf(w, "  type:       %s\n", a.Type)
	if len(a.Roles) > 0 {
		fmt.Fprintf(w, "  roles:      %v\n", a.Roles)
	}
	if a.Runtime != "" {
		fmt.Fprintf(w, "  runtime:    %s\n", a.Runtime)
	}
	if a.Model != "" {
		fmt.Fprintf(w, "  model:      %s\n", a.Model)
	}
	if a.Sandbox != "" {
		fmt.Fprintf(w, "  sandbox:    %s\n", a.Sandbox)
	}
	if len(a.Capabilities.WorkTypes) > 0 {
		fmt.Fprintf(w, "  work_types: %v\n", a.Capabilities.WorkTypes)
	}
	return nil
}

func runActorValidate(ctx *Context, args []string) error {
	fs := ctx.NewFlagSet("gw actor validate")
	if _, err := parseFlags(fs, args); err != nil {
		return err
	}
	_, reg, warnings, err := ctx.loadRegistry()
	if err != nil {
		return err
	}
	if ctx.JSON {
		return ctx.PrintJSON(map[string]any{"valid": true, "actors": len(reg.Actors), "warnings": warnings})
	}
	for _, w := range warnings {
		fmt.Fprintf(ctx.Stdout, "warning: %s\n", w)
	}
	fmt.Fprintf(ctx.Stdout, "actors.yaml is valid (%d actors)\n", len(reg.Actors))
	return nil
}
