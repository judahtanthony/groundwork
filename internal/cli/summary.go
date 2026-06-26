package cli

import (
	"fmt"
	"time"

	"groundwork/internal/completion"
)

// newTicketSummaryCmd is `gw ticket summary` (ADR 0047/0057): record and show a
// node's completion summary, the unit the bulk review bundle aggregates.
func newTicketSummaryCmd() *Command {
	return &Command{
		Name:  "summary",
		Usage: "Record or show a node's completion summary",
		Sub: []*Command{
			{Name: "set", Usage: "Record a completion summary", Args: "<id>", Run: runSummarySet, Flags: []FlagDoc{
				{"--outcome", "one-line outcome (required)"},
				{"--changed", "changed file (repeatable)"},
				{"--decision", "decision (repeatable)"},
				{"--assumption", "assumption (repeatable)"},
				{"--risk", "risk (repeatable)"},
				{"--canon", "canon update (repeatable)"},
			}},
			{Name: "show", Usage: "Show a node's completion summary", Args: "<id>", Run: runSummaryShow},
		},
	}
}

func runSummarySet(ctx *Context, args []string) error {
	fs := ctx.NewFlagSet("gw ticket summary set")
	var outcome string
	var changed, decisions, assumptions, risks, canon stringSlice
	fs.StringVar(&outcome, "outcome", "", "one-line outcome")
	fs.Var(&changed, "changed", "changed file")
	fs.Var(&decisions, "decision", "decision")
	fs.Var(&assumptions, "assumption", "assumption")
	fs.Var(&risks, "risk", "risk")
	fs.Var(&canon, "canon", "canon update")
	pos, err := parseFlags(fs, args)
	if err != nil {
		return err
	}
	if len(pos) != 1 || outcome == "" {
		return &Error{Code: "invalid_args", Message: "usage: gw ticket summary set <id> --outcome <text> [--changed f ...]"}
	}
	proj, db, err := ctx.openStore()
	if err != nil {
		return err
	}
	defer db.Close()
	if _, err := db.GetTicket(pos[0]); err != nil {
		return &Error{Code: "not_found", Message: "ticket not found"}
	}
	s := &completion.Summary{
		NodeID: pos[0], Outcome: outcome, Changed: changed,
		Decisions: decisions, Assumptions: assumptions, Risks: risks, CanonUpdates: canon,
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}
	if err := completion.Write(proj.TicketsDir(), s); err != nil {
		return &Error{Code: "sidecar_error", Message: err.Error()}
	}
	if err := db.UpsertCompletionSummary(s); err != nil {
		return &Error{Code: "store_error", Message: err.Error()}
	}
	if ctx.JSON {
		return ctx.PrintJSON(s)
	}
	fmt.Fprintf(ctx.Stdout, "Recorded completion summary for %s.\n", s.NodeID)
	return nil
}

func runSummaryShow(ctx *Context, args []string) error {
	fs := ctx.NewFlagSet("gw ticket summary show")
	pos, err := parseFlags(fs, args)
	if err != nil {
		return err
	}
	if len(pos) != 1 {
		return &Error{Code: "invalid_args", Message: "usage: gw ticket summary show <id>"}
	}
	proj, db, err := ctx.openStore()
	if err != nil {
		return err
	}
	defer db.Close()
	s, err := db.GetCompletionSummary(pos[0])
	if err != nil {
		return &Error{Code: "store_error", Message: err.Error()}
	}
	if s == nil {
		// Fall back to the authoritative sidecar.
		if sc, ok, _ := completion.Read(proj.TicketsDir(), pos[0]); ok {
			s = sc
		}
	}
	if s == nil {
		if ctx.JSON {
			return ctx.PrintJSON(nil)
		}
		fmt.Fprintf(ctx.Stdout, "No completion summary for %s.\n", pos[0])
		return nil
	}
	if ctx.JSON {
		return ctx.PrintJSON(s)
	}
	fmt.Fprintf(ctx.Stdout, "Completion summary %s\n  outcome: %s\n", s.NodeID, s.Outcome)
	printList(ctx, "changed", s.Changed)
	printList(ctx, "decisions", s.Decisions)
	printList(ctx, "assumptions", s.Assumptions)
	printList(ctx, "risks", s.Risks)
	printList(ctx, "canon", s.CanonUpdates)
	return nil
}

func printList(ctx *Context, label string, items []string) {
	for _, it := range items {
		fmt.Fprintf(ctx.Stdout, "  %s: %s\n", label, it)
	}
}
