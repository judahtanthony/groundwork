package cli

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

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
	Reparent(id, newParentID, actor string) error
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
	// Route to the coordinator only when it serves THIS project root. A coordinator
	// running for another project (or a stale one on the default port) must not
	// silently capture our mutations; on mismatch we fall back to the direct store
	// (T-1033).
	if c := client.New(p.Config.Server.Addr); coordinatorServes(c, p.Root) {
		return c, func() {}, nil
	}
	db, err := openDB(p)
	if err != nil {
		return nil, nil, err
	}
	// Enable filesystem write-through (ADR 0053) so offline mutations rewrite their
	// ticket sidecars, keeping files the source of truth — the same contract
	// openStore honors. Without this, a direct-store transition/create updates
	// SQLite but not ticket.md, and the next `gw server` boot flags a spurious
	// recovery_needed divergence.
	db.SetExportDir(p.TicketsDir())
	return db, func() { db.Close() }, nil
}

// coordinatorServes reports whether a reachable coordinator at c is serving the
// given project root (T-1033).
func coordinatorServes(c *client.Client, root string) bool {
	served, ok := c.CoordinatorRoot()
	return ok && sameRoot(served, root)
}

// sameRoot compares two project roots for equality, tolerating path formatting
// differences. Both come from config discovery, so a cleaned compare suffices.
func sameRoot(a, b string) bool {
	return filepath.Clean(a) == filepath.Clean(b)
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
	served, ok := c.CoordinatorRoot()
	if !ok {
		return nil, &Error{
			Code:    "coordinator_required",
			Message: "this command requires a running coordinator; start it with \"gw server\"",
		}
	}
	// A coordinator serving a different project must not land or decide approvals
	// for this one (T-1033).
	if !sameRoot(served, p.Root) {
		return nil, &Error{
			Code:    "coordinator_mismatch",
			Message: fmt.Sprintf("the coordinator on %s serves a different project (%s); start gw server in this project", p.Config.Server.Addr, served),
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
