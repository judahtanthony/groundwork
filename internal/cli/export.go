package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"groundwork/internal/config"
	"groundwork/internal/exporter"
	"groundwork/internal/store/sqlite"
	"groundwork/internal/ticket"
)

// newExportCmd is the top-level `gw export` (all tickets).
func newExportCmd() *Command {
	return &Command{Name: "export", Usage: "Export tickets to Markdown", Run: runExport}
}

// newTicketExportCmd is `gw ticket export [id]`.
func newTicketExportCmd() *Command {
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

	var written []string
	for _, t := range tickets {
		path, err := exportTicket(p, db, t)
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

// exportTicket renders t and writes .groundwork/tickets/<id>/ticket.md.
func exportTicket(p *config.Project, db *sqlite.DB, t *ticket.Ticket) (string, error) {
	deps, err := db.DependencyIDs(t.ID)
	if err != nil {
		return "", err
	}
	data, err := exporter.Render(t, deps)
	if err != nil {
		return "", err
	}
	dir := filepath.Join(p.TicketsDir(), t.ID)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	path := filepath.Join(dir, "ticket.md")
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return "", err
	}
	return path, nil
}
