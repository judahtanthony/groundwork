package sqlite

import (
	"bytes"
	"os"
	"path/filepath"

	"groundwork/internal/decision"
	"groundwork/internal/encoding"
	"groundwork/internal/exporter"
)

// DivergenceReport lists tickets whose live SQLite state does not match their
// committed sidecar — i.e. a durable mutation that hit SQLite but was never
// exported (a crash between the SQLite commit and the file write).
type DivergenceReport struct {
	Diverged []string `json:"diverged"` // ticket ids whose sidecar is missing or stale
}

// DetectFileDivergence compares each ticket's canonical export against its
// committed ticket.md sidecar (ADR 0053). A mismatch means SQLite holds an
// unexported durable mutation. Rather than silently treating SQLite as newer
// truth (the normal repair path is to rebuild SQLite from files), it surfaces a
// recovery_needed record appended to the node's sidecar — non-destructively,
// without overwriting the diverged ticket.md from SQLite — so the lost-or-newer
// mutation is visible and a human/recovery decides.
//
// It is a no-op when write-through is disabled (no authoritative directory) and
// is only meaningful when SQLite survived a restart (a freshly rebuilt store
// matches its files by construction).
func (db *DB) DetectFileDivergence() (*DivergenceReport, error) {
	rep := &DivergenceReport{}
	if db.exportDir == "" {
		return rep, nil
	}
	tickets, err := db.ListTickets()
	if err != nil {
		return nil, err
	}
	depMap, err := db.DependencyMap()
	if err != nil {
		return nil, err
	}
	for _, t := range tickets {
		want, err := exporter.Render(t, depMap[t.ID])
		if err != nil {
			return nil, err
		}
		got, err := os.ReadFile(filepath.Join(db.exportDir, t.ID, "ticket.md"))
		if err != nil && !os.IsNotExist(err) {
			return nil, err
		}
		if err == nil && bytes.Equal(got, want) {
			continue // sidecar matches store: consistent
		}
		rep.Diverged = append(rep.Diverged, t.ID)
		if err := appendRecoveryNeededToSidecar(db.exportDir, t.ID); err != nil {
			return nil, err
		}
	}
	return rep, nil
}

// appendRecoveryNeededToSidecar appends a recovery_needed record to a node's
// decisions.ndjson on disk without touching ticket.md, so divergence is flagged
// durably without overwriting the diverged ticket from SQLite. Idempotent: it
// does not append a second record when a pending recovery_needed already exists.
func appendRecoveryNeededToSidecar(dir, ticketID string) error {
	recs, _, err := decision.Read(dir, ticketID)
	if err != nil {
		return err
	}
	for _, r := range recs {
		if r.EventType == decision.EventRecoveryNeeded && r.Status == decision.StatusPending {
			return nil // already flagged
		}
	}
	seq := 0
	for _, r := range recs {
		if r.Sequence > seq {
			seq = r.Sequence
		}
	}
	recs = append(recs, decision.Record{
		Sequence: seq + 1, EventType: decision.EventRecoveryNeeded, TicketID: ticketID,
		Status: decision.StatusPending, RequestedAt: encoding.Now(),
		HandoffSummary: "recovery_needed: SQLite state diverged from the committed sidecar (unexported durable mutation); rebuild from files to repair",
	})
	return decision.Write(dir, ticketID, recs)
}
