package cli

import (
	"fmt"
)

// newReviewCmd is `gw review` (ADR 0057): assemble a feature-level review bundle
// so a human reviews a composite/root once at the boundary instead of per child.
func newReviewCmd() *Command {
	return &Command{
		Name:  "review",
		Usage: "Feature-level review of a composite/root subtree",
		Sub: []*Command{
			{Name: "bundle", Usage: "Assemble the review bundle for a node", Args: "<id>", Run: runReviewBundle},
		},
	}
}

func runReviewBundle(ctx *Context, args []string) error {
	fs := ctx.NewFlagSet("gw review bundle")
	pos, err := parseFlags(fs, args)
	if err != nil {
		return err
	}
	if len(pos) != 1 {
		return &Error{Code: "invalid_args", Message: "usage: gw review bundle <id>"}
	}
	_, db, err := ctx.openStore()
	if err != nil {
		return err
	}
	defer db.Close()
	b, err := db.ReviewBundle(pos[0])
	if err != nil {
		return &Error{Code: "bundle_failed", Message: err.Error()}
	}
	if ctx.JSON {
		return ctx.PrintJSON(b)
	}
	fmt.Fprintf(ctx.Stdout, "Review bundle %s — %s\n", b.NodeID, b.Title)
	fmt.Fprintf(ctx.Stdout, "  recommendation: %s\n", b.Recommendation)
	if len(b.UnresolvedExceptions) > 0 {
		fmt.Fprintf(ctx.Stdout, "  unresolved exceptions: %v\n", b.UnresolvedExceptions)
	}
	for _, c := range b.Children {
		outcome := "(no summary)"
		if c.Summary != nil {
			outcome = c.Summary.Outcome
		}
		fmt.Fprintf(ctx.Stdout, "  - %s [%s] %s — %s\n", c.NodeID, c.Status, c.Title, outcome)
		for _, v := range c.Validations {
			fmt.Fprintf(ctx.Stdout, "      validation %s: %s\n", v.Name, v.Status)
		}
		if len(c.Exceptions) > 0 {
			fmt.Fprintf(ctx.Stdout, "      exceptions: %v\n", c.Exceptions)
		}
	}
	return nil
}
