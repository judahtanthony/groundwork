package cli

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"groundwork/internal/decision"
	"groundwork/internal/exporter"
	"groundwork/internal/store/sqlite"
	"groundwork/internal/ticket"
)

func newTicketImportCmd() *Command {
	return &Command{Name: "import", Usage: "Rebuild ticket rows from committed Markdown exports", Args: "[path]", Run: runTicketImport}
}

func runTicketImport(ctx *Context, args []string) error {
	fs := ctx.NewFlagSet("gw ticket import")
	pos, err := parseFlags(fs, args)
	if err != nil {
		return err
	}
	p, db, err := ctx.openStore()
	if err != nil {
		return err
	}
	defer db.Close()

	dir := p.TicketsDir()
	if len(pos) > 0 {
		dir = pos[0]
	}
	n, err := importExports(db, dir)
	if err != nil {
		return &Error{Code: "import_failed", Message: err.Error()}
	}
	if ctx.JSON {
		return ctx.PrintJSON(map[string]int{"imported": n})
	}
	fmt.Fprintf(ctx.Stdout, "Imported %d ticket(s) from %s\n", n, dir)
	return nil
}

// importExports parses every ticket.md under dir and rebuilds node rows and
// dependency edges, preserving ids and timestamps (T-0902). Existing nodes are
// skipped (idempotent). Runtime state (leases/runs) is not restored. Nodes are
// inserted parent-before-child to satisfy the foreign key.
func importExports(db *sqlite.DB, dir string) (int, error) {
	if _, err := os.Stat(dir); err != nil {
		if os.IsNotExist(err) {
			return 0, nil
		}
		return 0, err
	}

	type entry struct {
		t    *ticket.Ticket
		deps []string
	}
	var entries []entry
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || d.Name() != "ticket.md" {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		t, deps, err := exporter.Parse(data)
		if err != nil {
			return fmt.Errorf("%s: %w", path, err)
		}
		entries = append(entries, entry{t: t, deps: deps})
		return nil
	})
	if err != nil {
		return 0, err
	}

	// Skip nodes that already exist (idempotent re-import).
	pending := entries[:0:0]
	maxSeq := 0
	for _, e := range entries {
		// Only T-NNNN ids feed the runtime allocator (ADR 0019, 0032). Goal/epic
		// ids (G-/E-) share the numeric-suffix shape but live outside the T
		// sequence; counting them could vault the allocator past unrelated numbers
		// (e.g. an E-2000 epic) and waste the T-id space.
		if strings.HasPrefix(e.t.ID, "T-") {
			if n := idSeq(e.t.ID); n > maxSeq {
				maxSeq = n
			}
		}
		if _, err := db.GetTicket(e.t.ID); err == nil {
			continue // already present
		} else if !errors.Is(err, sqlite.ErrNotFound) {
			return 0, err
		}
		pending = append(pending, e)
	}

	// Insert parent-before-child: repeatedly insert nodes whose parent is empty
	// or already present, until none remain.
	inserted := 0
	present := map[string]bool{}
	for len(pending) > 0 {
		progress := false
		next := pending[:0]
		for _, e := range pending {
			parentReady := e.t.ParentID == "" || present[e.t.ParentID] || ticketExists(db, e.t.ParentID)
			if !parentReady {
				next = append(next, e)
				continue
			}
			if err := db.ImportTicket(e.t); err != nil {
				return inserted, err
			}
			present[e.t.ID] = true
			inserted++
			progress = true
		}
		pending = next
		if !progress {
			return inserted, fmt.Errorf("import: %d node(s) have unresolved parents", len(pending))
		}
	}

	// Keep id allocation past the imported maximum (ADR 0019).
	if maxSeq > 0 {
		if err := db.SeedTicketSeq(maxSeq); err != nil {
			return inserted, err
		}
	}

	// Dependency edges, now that all nodes exist. AddDependency is idempotent on a
	// duplicate edge (re-import), so the only benign error is a missing endpoint
	// (a depends_on target absent from the export set) — skip that edge. Every
	// other error (cycle, self-edge, or a real store failure) is a genuine import
	// problem and must fail loudly rather than silently drop edges.
	for _, e := range entries {
		for _, dep := range e.deps {
			if err := db.AddDependency(e.t.ID, dep, ownerActor); err != nil {
				if errors.Is(err, sqlite.ErrNotFound) {
					continue // missing endpoint: tolerate, edge skipped
				}
				return inserted, fmt.Errorf("add dependency %s -> %s: %w", e.t.ID, dep, err)
			}
		}
	}

	// Durable decision records: rebuild the live projection from each ticket's
	// decisions.ndjson sidecar (ADR 0051/0053). Sequences are preserved so a
	// rebuilt store re-exports byte-for-byte.
	for _, e := range entries {
		recs, ok, err := decision.Read(dir, e.t.ID)
		if err != nil {
			return inserted, fmt.Errorf("read decisions %s: %w", e.t.ID, err)
		}
		if !ok {
			continue
		}
		for _, rec := range recs {
			if err := db.ImportDecision(rec); err != nil {
				return inserted, fmt.Errorf("import decision %s seq %d: %w", e.t.ID, rec.Sequence, err)
			}
		}
	}
	return inserted, nil
}

func ticketExists(db *sqlite.DB, id string) bool {
	_, err := db.GetTicket(id)
	return err == nil
}

// idSeq extracts the numeric suffix of an id like "T-0007" -> 7, or 0.
func idSeq(id string) int {
	i := strings.LastIndex(id, "-")
	if i < 0 {
		return 0
	}
	n, err := strconv.Atoi(id[i+1:])
	if err != nil {
		return 0
	}
	return n
}
