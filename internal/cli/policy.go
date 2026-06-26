package cli

import (
	"errors"
	"fmt"

	"groundwork/internal/store/sqlite"
)

// suggestionError maps a store error from a suggestion decision to a CLI error.
func suggestionError(err error) error {
	if errors.Is(err, sqlite.ErrNotFound) {
		return &Error{Code: "not_found", Message: "suggestion not found"}
	}
	return &Error{Code: "store_error", Message: err.Error()}
}

// newPolicyCmd is the `gw policy` group. v1 surfaces the elevation-readiness
// suggestion queue (ADR 0038): the system proposes autonomy elevations for human
// review and never self-elevates.
func newPolicyCmd() *Command {
	return &Command{
		Name:  "policy",
		Usage: "Inspect policy and elevation suggestions",
		Sub: []*Command{
			{Name: "suggestions", Usage: "List elevation-readiness suggestions", Run: runPolicySuggestions, Flags: []FlagDoc{
				{"--scan", "scan the work tree for new elevation candidates first"},
				{"--all", "include promoted and dismissed suggestions"},
			}},
			{Name: "promote", Usage: "Mark a suggestion promoted and print the policy change to apply", Args: "<id>", Run: runPolicyPromote},
			{Name: "dismiss", Usage: "Dismiss a suggestion", Args: "<id>", Run: runPolicyDismiss},
		},
	}
}

func runPolicySuggestions(ctx *Context, args []string) error {
	fs := ctx.NewFlagSet("gw policy suggestions")
	var scan, all bool
	fs.BoolVar(&scan, "scan", false, "scan for new candidates first")
	fs.BoolVar(&all, "all", false, "include promoted and dismissed")
	if err := fs.Parse(args); err != nil {
		return err
	}
	_, db, err := ctx.openStore()
	if err != nil {
		return err
	}
	defer db.Close()

	if scan {
		if _, err := db.GenerateElevationSuggestions(); err != nil {
			return &Error{Code: "scan_failed", Message: err.Error()}
		}
	}
	status := "pending"
	if all {
		status = ""
	}
	suggestions, err := db.ListSuggestions(status)
	if err != nil {
		return &Error{Code: "list_failed", Message: err.Error()}
	}
	if ctx.JSON {
		return ctx.PrintJSON(suggestions)
	}
	if len(suggestions) == 0 {
		fmt.Fprintln(ctx.Stdout, "No elevation suggestions.")
		return nil
	}
	fmt.Fprintf(ctx.Stdout, "%-8s  %-10s  %-22s  %-6s  %s\n", "ID", "STATUS", "TARGET", "LEVEL", "RATIONALE")
	for _, s := range suggestions {
		target := fmt.Sprintf("%s@%s", s.ActionType, s.WorkType)
		fmt.Fprintf(ctx.Stdout, "%-8s  %-10s  %-22s  %-6s  %s\n", s.ID, s.Status, target, s.Level, s.Rationale)
	}
	return nil
}

func runPolicyPromote(ctx *Context, args []string) error {
	if len(args) != 1 {
		return &Error{Code: "invalid_args", Message: "usage: gw policy promote <id>"}
	}
	_, db, err := ctx.openStore()
	if err != nil {
		return err
	}
	defer db.Close()
	s, err := db.SetSuggestionStatus(args[0], "promoted")
	if err != nil {
		return suggestionError(err)
	}
	if ctx.JSON {
		return ctx.PrintJSON(s)
	}
	// Never self-elevate: emit the change for a human to apply. Amending policy is
	// itself a human-gated action (ADR 0038).
	fmt.Fprintf(ctx.Stdout, "Promoted %s. Apply this to .groundwork/policies/autonomy.yaml by hand:\n\n", s.ID)
	fmt.Fprintf(ctx.Stdout, "actions:\n  %s:\n    by_work_type:\n      %s:\n        level: %s\n",
		s.ActionType, s.WorkType, s.Level)
	return nil
}

func runPolicyDismiss(ctx *Context, args []string) error {
	if len(args) != 1 {
		return &Error{Code: "invalid_args", Message: "usage: gw policy dismiss <id>"}
	}
	_, db, err := ctx.openStore()
	if err != nil {
		return err
	}
	defer db.Close()
	s, err := db.SetSuggestionStatus(args[0], "dismissed")
	if err != nil {
		return suggestionError(err)
	}
	if ctx.JSON {
		return ctx.PrintJSON(s)
	}
	fmt.Fprintf(ctx.Stdout, "Dismissed %s.\n", s.ID)
	return nil
}
