package cli

import (
	"fmt"
	"path/filepath"

	"groundwork/internal/exporter"
	"groundwork/internal/ticket"
)

// newExportCmd backs both `gw export` and `gw ticket export [id]`.
func newExportCmd() *Command {
	return &Command{Name: "export", Usage: "Export tickets to Markdown", Args: "[id]", Run: runExport}
}

func runExport(ctx *Context, args []string) error {
	fs := ctx.NewFlagSet("gw export")
	pos, err := parseFlags(fs, args)
	if err != nil {
		return err
	}

	p, db, err := ctx.openStore()
	if err != nil {
		return err
	}
	defer db.Close()

	var tickets []*ticket.Ticket
	if len(pos) >= 1 {
		t, err := db.GetTicket(pos[0])
		if err != nil {
			return ticketError(err, pos[0])
		}
		tickets = []*ticket.Ticket{t}
	} else {
		tickets, err = db.ListTickets()
		if err != nil {
			return &Error{Code: "list_failed", Message: err.Error()}
		}
	}

	// One bulk dependency query instead of one per ticket.
	depMap, err := db.DependencyMap()
	if err != nil {
		return &Error{Code: "store_error", Message: err.Error()}
	}

	var written []string
	for _, t := range tickets {
		path, err := exporter.WriteTo(p.TicketsDir(), t, depMap[t.ID])
		if err != nil {
			return &Error{Code: "export_failed", Message: err.Error()}
		}
		rel, _ := filepath.Rel(p.Root, path)
		written = append(written, rel)
	}

	if ctx.JSON {
		if written == nil {
			written = []string{}
		}
		return ctx.PrintJSON(map[string]any{"exported": written})
	}
	if len(written) == 0 {
		fmt.Fprintln(ctx.Stdout, "No tickets to export.")
		return nil
	}
	for _, w := range written {
		fmt.Fprintf(ctx.Stdout, "exported %s\n", w)
	}
	return nil
}
