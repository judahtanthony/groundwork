package server

import (
	"fmt"
	"net/http"
	"strings"

	"groundwork/internal/exporter"
)

// completeLanding finishes a landing whose store transition to done has already
// happened: it records the ratification and makes the durable git commit
// (ADR 0034). On git failure it writes a 500 envelope and returns false — the
// node is recorded done but uncommitted, and re-running `gw ticket land <id>`
// finishes the commit idempotently (handleTicketLand routes an already-done node
// straight back here). Returns true when the landing is fully recorded: commit
// made, nothing to commit, or no git work tree.
func (s *Server) completeLanding(w http.ResponseWriter, id, reason string) bool {
	if err := s.finishLanding(id, reason); err != nil {
		writeError(w, http.StatusInternalServerError, "land_commit_failed", fmt.Sprintf(
			"%s is recorded landed but the git commit failed: %v; resolve the git issue and "+
				"run \"gw ticket land %s\" to finish the commit", id, err, id))
		return false
	}
	return true
}

// finishLanding records the ratification and makes the durable git commit for a
// landing whose store transition to done has already happened, returning any
// commit error so non-HTTP callers (the decision recorder) can map it onto their
// own response surface. completeLanding wraps this for the JSON handlers.
func (s *Server) finishLanding(id, reason string) error {
	s.ratify(id, "land", reason)
	if err := s.commitLanding(id); err != nil {
		return err
	}
	// For a root with an integration branch, landing to main merges that branch
	// and cleans it up (ADR 0058); a no-op for ordinary nodes.
	return s.mergeRootToMain(id)
}

// commitLanding regenerates the node's Markdown export (now status=done) and, in a
// git work tree, commits it on the current branch alongside whatever the human
// staged for this ticket (ADR 0034). The git index is the ticket-scoped pathspec:
// it is the human's explicit selection of what belongs to this landing.
//
// The export is rewritten even when there is no git work tree (s.repo == nil), so
// the durable export never drifts from store status in non-git projects; only the
// commit is skipped. Commits are refused on a detached HEAD (the commit would be
// an orphan). The commit SHA is recorded on the audit trail.
func (s *Server) commitLanding(ticketID string) error {
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
	if s.repo == nil {
		return nil // no git work tree: export refreshed, nothing to commit
	}
	// Refuse to commit on a detached HEAD: the commit would be unreachable once
	// HEAD moves, silently losing the landing.
	branch, err := s.repo.CurrentBranch()
	if err != nil {
		return err
	}
	if branch == "HEAD" {
		return fmt.Errorf("refusing to land on a detached HEAD; check out a branch and retry")
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
