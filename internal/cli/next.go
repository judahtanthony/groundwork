package cli

import (
	"fmt"

	"groundwork/internal/contextbrief"
	"groundwork/internal/store/sqlite"
	"groundwork/internal/ticket"
)

// nextNode returns the top eligible node in ADR 0039 value order, or nil when
// nothing is ready. It is the pick `gw next` recommends and `--claim` takes.
func nextNode(db *sqlite.DB) (*ticket.Ticket, error) {
	eligible, err := db.ListEligibleOrdered()
	if err != nil {
		return nil, err
	}
	if len(eligible) == 0 {
		return nil, nil
	}
	return eligible[0], nil
}

// newNextCmd is `gw next`: the human picker over the same eligible, value-ordered
// set the scheduler dispatches from (ADR 0039/0041). It recommends the top node
// and, with --claim, takes it.
func newNextCmd() *Command {
	return &Command{Name: "next", Usage: "Show the next eligible node to work on (top of the ready set)", Run: runNext, Flags: []FlagDoc{
		{"--claim", "claim the top node: assign it and start work"},
		{"--actor <id>", "assignee when --claim is set (default: human.owner)"},
	}}
}

func runNext(ctx *Context, args []string) error {
	fs := ctx.NewFlagSet("gw next")
	var claim bool
	var assignee string
	fs.BoolVar(&claim, "claim", false, "claim the top node: assign it and start work")
	fs.StringVar(&assignee, "actor", ownerActor, "assignee when --claim is set (default: human.owner)")
	if err := fs.Parse(args); err != nil {
		return err
	}

	p, db, err := ctx.openStore()
	if err != nil {
		return err
	}
	defer db.Close()

	top, err := nextNode(db)
	if err != nil {
		return &Error{Code: "list_failed", Message: err.Error()}
	}
	if top == nil {
		if ctx.JSON {
			return ctx.PrintJSON(map[string]any{"next": nil})
		}
		fmt.Fprintln(ctx.Stdout, "No eligible nodes. Nothing is ready to work on.")
		return nil
	}

	// --claim takes the top node through the same guarded path as gw ticket claim.
	if claim {
		store, closeStore, err := ctx.openTicketStore()
		if err != nil {
			return err
		}
		defer closeStore()
		if err := claimNode(db, store, top.ID, assignee); err != nil {
			return err
		}
		if ctx.JSON {
			return ctx.PrintJSON(claimedJSON(top.ID, assignee))
		}
		printClaimed(ctx, p, db, top.ID, assignee)
		return nil
	}

	if ctx.JSON {
		deps, err := db.DependencyIDs(top.ID)
		if err != nil {
			return &Error{Code: "store_error", Message: err.Error()}
		}
		return ctx.PrintJSON(struct {
			*ticket.Ticket
			DependsOn []string `json:"depends_on"`
		}{top, deps})
	}

	fmt.Fprintf(ctx.Stdout, "Next: %s  %s\n\n", top.ID, top.Title)
	if brief, err := contextbrief.Build(db, p, top.ID, false); err == nil {
		renderBrief(ctx, brief)
	}
	fmt.Fprintf(ctx.Stdout, "\nTake it: gw ticket claim %s   (or: gw next --claim)\n", top.ID)
	return nil
}
