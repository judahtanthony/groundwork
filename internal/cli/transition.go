package cli

import (
	"errors"
	"fmt"

	"groundwork/internal/store/sqlite"
	"groundwork/internal/ticket"
)

func newTicketTransitionCmd() *Command {
	return &Command{Name: "transition", Usage: "Transition a node's status", Args: "<id> <status>", Run: runTicketTransition}
}

func runTicketTransition(ctx *Context, args []string) error {
	fs := ctx.NewFlagSet("gw ticket transition")
	pos, err := parseFlags(fs, args)
	if err != nil {
		return err
	}
	if len(pos) < 2 {
		return &Error{Code: "invalid_args", Message: "usage: gw ticket transition <id> <status>"}
	}
	id := pos[0]
	to := ticket.Status(pos[1])

	store, closeStore, err := ctx.openTicketStore()
	if err != nil {
		return err
	}
	defer closeStore()

	if err := store.TransitionTicket(id, to, ownerActor); err != nil {
		switch {
		case errors.Is(err, sqlite.ErrNotFound):
			return ticketError(err, id)
		case errors.Is(err, sqlite.ErrIllegalTransition):
			return &Error{Code: "illegal_transition", Message: err.Error()}
		default:
			return &Error{Code: "transition_failed", Message: err.Error()}
		}
	}

	if ctx.JSON {
		return ctx.PrintJSON(map[string]string{"id": id, "status": string(to)})
	}
	fmt.Fprintf(ctx.Stdout, "%s -> %s\n", id, to)
	return nil
}

func newTicketTriageCmd() *Command {
	return &Command{Name: "triage", Usage: "Classify a node as leaf or composite", Args: "<id> <leaf|composite>", Run: runTicketTriage}
}

func runTicketTriage(ctx *Context, args []string) error {
	fs := ctx.NewFlagSet("gw ticket triage")
	pos, err := parseFlags(fs, args)
	if err != nil {
		return err
	}
	if len(pos) < 2 {
		return &Error{Code: "invalid_args", Message: "usage: gw ticket triage <id> <leaf|composite>"}
	}
	id := pos[0]
	nt := ticket.NodeType(pos[1])

	_, db, err := ctx.openStore()
	if err != nil {
		return err
	}
	defer db.Close()

	if err := db.TriageTicket(id, nt, ownerActor); err != nil {
		if errors.Is(err, sqlite.ErrNotFound) {
			return ticketError(err, id)
		}
		return &Error{Code: "triage_failed", Message: err.Error()}
	}

	if ctx.JSON {
		return ctx.PrintJSON(map[string]string{"id": id, "node_type": string(nt)})
	}
	fmt.Fprintf(ctx.Stdout, "%s triaged as %s\n", id, nt)
	return nil
}
