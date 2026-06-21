package server

import (
	"fmt"
	"strings"

	"groundwork/internal/exporter"
)

// commitLanding makes the durable git commit for a landed node (ADR 0034). The
// store-side land has already transitioned the node to done; this regenerates
// the node's Markdown export (now status=done), stages it alongside whatever the
// human staged for this ticket, and commits on the current branch. The git index
// is the ticket-scoped pathspec: it is the human's explicit selection of what
// belongs to this landing, so unrelated unstaged edits are never captured.
//
// It is best-effort by environment: when the project root is not a git work tree
// (s.repo == nil) the landing is still recorded in the store and this is a no-op.
// A genuine git failure is returned so the caller can surface it; the commit SHA
// is recorded on the audit trail (T-1004).
func (s *Server) commitLanding(ticketID string) error {
	if s.repo == nil {
		return nil
	}
	t, err := s.db.GetTicket(ticketID)
	if err != nil {
		return err
	}
	deps, err := s.db.DependencyIDs(ticketID)
	if err != nil {
		return err
	}
	path, err := exporter.WriteTo(s.proj.TicketsDir(), t, deps)
	if err != nil {
		return fmt.Errorf("regenerate export: %w", err)
	}
	if err := s.repo.Add(path); err != nil {
		return err
	}
	staged, err := s.repo.HasStagedChanges()
	if err != nil {
		return err
	}
	if !staged {
		// No work staged and the export was already current: record the landing
		// without forcing an empty commit.
		s.ratify(ticketID, "land", "landed; no staged changes to commit")
		return nil
	}
	sha, err := s.repo.Commit(fmt.Sprintf("Land %s: %s", t.ID, strings.TrimSpace(t.Title)))
	if err != nil {
		return err
	}
	_ = s.db.AppendAuditEvent(ownerActor, "ticket.committed", "ticket", ticketID, map[string]any{
		"sha":  sha,
		"path": path,
	})
	s.ratify(ticketID, "land", "committed "+sha)
	return nil
}
