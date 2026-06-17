package cli

import (
	"errors"
	"fmt"

	"groundwork/internal/store/sqlite"
)

func newTicketLinkCmd() *Command {
	return &Command{Name: "link", Usage: "Add or remove a dependency edge", Args: "<id> --depends-on <id>", Run: runTicketLink}
}

func runTicketLink(ctx *Context, args []string) error {
	fs := ctx.NewFlagSet("gw ticket link")
	var dependsOn string
	var remove bool
	fs.StringVar(&dependsOn, "depends-on", "", "id this node depends on (required)")
	fs.BoolVar(&remove, "remove", false, "remove the edge instead of adding it")
	pos, err := parseFlags(fs, args)
	if err != nil {
		return err
	}
	if len(pos) < 1 || dependsOn == "" {
		return &Error{Code: "invalid_args", Message: "usage: gw ticket link <id> --depends-on <id> [--remove]"}
	}
	id := pos[0]

	_, db, err := ctx.openStore()
	if err != nil {
		return err
	}
	defer db.Close()

	if remove {
		if err := db.RemoveDependency(id, dependsOn, actor); err != nil {
			if errors.Is(err, sqlite.ErrNotFound) {
				return &Error{Code: "not_found", Message: fmt.Sprintf("no edge %s -> %s", id, dependsOn)}
			}
			return &Error{Code: "link_failed", Message: err.Error()}
		}
		return linkOutput(ctx, id, dependsOn, false)
	}

	if err := db.AddDependency(id, dependsOn, actor); err != nil {
		switch {
		case errors.Is(err, sqlite.ErrSelfDependency):
			return &Error{Code: "self_dependency", Message: err.Error()}
		case errors.Is(err, sqlite.ErrDependencyCycle):
			return &Error{Code: "dependency_cycle", Message: fmt.Sprintf("%s: %s -> %s", err.Error(), id, dependsOn)}
		case errors.Is(err, sqlite.ErrNotFound):
			return &Error{Code: "not_found", Message: "one or both nodes do not exist"}
		default:
			return &Error{Code: "link_failed", Message: err.Error()}
		}
	}
	return linkOutput(ctx, id, dependsOn, true)
}

func linkOutput(ctx *Context, id, dependsOn string, added bool) error {
	if ctx.JSON {
		return ctx.PrintJSON(map[string]any{"id": id, "depends_on": dependsOn, "added": added})
	}
	if added {
		fmt.Fprintf(ctx.Stdout, "%s now depends on %s\n", id, dependsOn)
	} else {
		fmt.Fprintf(ctx.Stdout, "%s no longer depends on %s\n", id, dependsOn)
	}
	return nil
}
