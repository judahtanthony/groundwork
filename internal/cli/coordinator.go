package cli

import (
	"errors"
	"os"

	"groundwork/internal/client"
	"groundwork/internal/config"
	"groundwork/internal/store/sqlite"
	"groundwork/internal/ticket"
)

// ticketStore is the transport-agnostic surface the mutating ticket commands
// use. Both *sqlite.DB (direct store) and *client.Client (coordinator API)
// satisfy it, so a command runs identically regardless of which is selected
// (ADR 0031). The client maps API error codes back to the store sentinels, so
// callers may branch on errors.Is(err, sqlite.Err…) either way.
type ticketStore interface {
	GetTicket(id string) (*ticket.Ticket, error)
	CreateTicket(t *ticket.Ticket, actor string) error
	UpdateTicket(t *ticket.Ticket, actor string) error
	TransitionTicket(id string, to ticket.Status, actor string) error
	AddDependency(fromID, toID, actor string) error
	RemoveDependency(fromID, toID, actor string) error
}

// resolveProject discovers the project root and loads its config without opening
// the database. It surfaces config warnings on stderr (ADR 0021).
func (ctx *Context) resolveProject() (*config.Project, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, &Error{Code: "cwd_failed", Message: err.Error()}
	}
	p, err := config.Open(cwd, ctx.RootFlag)
	if err != nil {
		if errors.Is(err, config.ErrProjectNotFound) {
			return nil, &Error{Code: "no_project", Message: err.Error()}
		}
		return nil, &Error{Code: "config_error", Message: err.Error()}
	}
	for _, w := range p.Warnings {
		if !ctx.JSON {
			ctx.Stderr.Write([]byte("gw: warning: " + w + "\n"))
		}
	}
	return p, nil
}

// openTicketStore selects the transport for a store-safe mutating command: it
// prefers the coordinator API when one is reachable (so the running server's
// state and SSE stream stay coherent), and otherwise falls back to the direct
// store (ADR 0031). The returned closer must be called when done.
func (ctx *Context) openTicketStore() (ticketStore, func(), error) {
	p, err := ctx.resolveProject()
	if err != nil {
		return nil, nil, err
	}
	if c := client.New(p.Config.Server.Addr); c.Healthy() {
		return c, func() {}, nil
	}
	db, err := openDB(p)
	if err != nil {
		return nil, nil, err
	}
	return db, func() { db.Close() }, nil
}

// requireCoordinator returns a client only when the coordinator is reachable.
// Live run-control commands use it so they fail clearly instead of mutating the
// store behind the coordinator's back (ADR 0031). Consumed by the run/approval
// commands as they land in later waves.
func (ctx *Context) requireCoordinator() (*client.Client, error) {
	p, err := ctx.resolveProject()
	if err != nil {
		return nil, err
	}
	c := client.New(p.Config.Server.Addr)
	if !c.Healthy() {
		return nil, &Error{
			Code:    "coordinator_required",
			Message: "this command requires a running coordinator; start it with \"gw server\"",
		}
	}
	return c, nil
}

// openDB opens and migrates the project's SQLite store.
func openDB(p *config.Project) (*sqlite.DB, error) {
	db, err := sqlite.Open(p.DBPath())
	if err != nil {
		return nil, &Error{Code: "store_error", Message: err.Error()}
	}
	if err := db.Migrate(); err != nil {
		db.Close()
		return nil, &Error{Code: "migrate_error", Message: err.Error()}
	}
	return db, nil
}
