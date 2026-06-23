package cli

import (
	"fmt"
	"strings"

	"groundwork/internal/canon"
	"groundwork/internal/contextbrief"
)

// newContextCmd is the `gw context <id>` command (also reused as
// `gw ticket context <id>`).
func newContextCmd() *Command {
	return &Command{Name: "context", Usage: "Show the bounded context brief for a node", Args: "<id>", Run: runContext, Flags: []FlagDoc{
		{"--siblings", "include sibling nodes (off by default)"},
		{"--miss <note>", "record a context-miss: something the brief lacked (ADR 0035)"},
	}}
}

func runContext(ctx *Context, args []string) error {
	fs := ctx.NewFlagSet("gw context")
	var siblings bool
	var miss string
	fs.BoolVar(&siblings, "siblings", false, "include sibling nodes (off by default)")
	fs.StringVar(&miss, "miss", "", "record a context-miss: something you needed that the brief lacked (ADR 0035)")
	pos, err := parseFlags(fs, args)
	if err != nil {
		return err
	}
	if len(pos) < 1 {
		return &Error{Code: "invalid_args", Message: "usage: gw context <id> [--siblings] [--miss <note>]"}
	}

	p, db, err := ctx.openStore()
	if err != nil {
		return err
	}
	defer db.Close()

	// --miss records what the brief failed to provide and returns. Misses are
	// ignored runtime state in the node journal; the review step promotes
	// recurring ones into canon (ADR 0035, ADR 0013).
	if strings.TrimSpace(miss) != "" {
		if _, err := db.GetTicket(pos[0]); err != nil {
			return ticketError(err, pos[0])
		}
		if err := canon.Miss(p.JournalDir(), pos[0], strings.TrimSpace(miss)); err != nil {
			return &Error{Code: "miss_failed", Message: err.Error()}
		}
		if ctx.JSON {
			return ctx.PrintJSON(map[string]string{"id": pos[0], "miss": strings.TrimSpace(miss)})
		}
		fmt.Fprintf(ctx.Stdout, "recorded context-miss for %s\n", pos[0])
		return nil
	}

	brief, err := contextbrief.Build(db, p, pos[0], siblings)
	if err != nil {
		return ticketError(err, pos[0])
	}

	if ctx.JSON {
		return ctx.PrintJSON(brief)
	}
	renderBrief(ctx, brief)
	return nil
}

func renderBrief(ctx *Context, b *contextbrief.Brief) {
	w := ctx.Stdout
	fmt.Fprintf(w, "Context for %s  %s\n", b.Node.ID, b.Node.Title)
	fmt.Fprintf(w, "  status: %s  type: %s\n\n", b.Node.Status, orDash(b.Node.NodeType, "untriaged"))

	fmt.Fprintln(w, "Acceptance criteria:")
	if len(b.Acceptance) == 0 {
		fmt.Fprintln(w, "  (none recorded)")
	} else {
		for _, a := range b.Acceptance {
			fmt.Fprintf(w, "  - %s\n", a)
		}
	}

	fmt.Fprintln(w, "\nAncestor spine:")
	if len(b.AncestorSpine) == 0 {
		fmt.Fprintln(w, "  (root node)")
	} else {
		for _, a := range b.AncestorSpine {
			fmt.Fprintf(w, "  %s  %s\n", a.ID, a.Title)
		}
	}

	fmt.Fprintln(w, "\nParent contract:")
	if b.ParentContract == "" {
		fmt.Fprintln(w, "  (none)")
	} else {
		fmt.Fprintf(w, "  %s\n", b.ParentContract)
	}

	fmt.Fprintln(w, "\nDependencies:")
	if len(b.Dependencies) == 0 {
		fmt.Fprintln(w, "  (none)")
	} else {
		for _, d := range b.Dependencies {
			fmt.Fprintf(w, "  %s  %s  [%s]\n", d.ID, d.Title, d.Status)
		}
	}

	fmt.Fprintln(w, "\nSOPs:")
	if len(b.SOPs) == 0 {
		fmt.Fprintln(w, "  (none)")
	} else {
		for _, s := range b.SOPs {
			fmt.Fprintf(w, "  %s\n", s)
		}
	}

	fmt.Fprintln(w, "\nOpen escalations:")
	if len(b.OpenEscalations) == 0 {
		fmt.Fprintln(w, "  (none)")
	} else {
		for _, e := range b.OpenEscalations {
			fmt.Fprintf(w, "  %s\n", e)
		}
	}

	if b.Siblings != nil {
		fmt.Fprintln(w, "\nSiblings:")
		if len(b.Siblings) == 0 {
			fmt.Fprintln(w, "  (none)")
		} else {
			for _, s := range b.Siblings {
				fmt.Fprintf(w, "  %s  %s  [%s]\n", s.ID, s.Title, s.Status)
			}
		}
	}
}
