package server

// Root integration-branch lifecycle (ADR 0058). On envelope approval the root
// gets a recorded integration target: in v1's single working tree that is the
// current feature branch (the workflow dogfooded by hand), or a freshly created
// gw/root/<id>-<slug> branch when starting from the default branch. Children land
// to this branch (land_to_parent); the gated root land_to_main merges and cleans
// it up. No worktrees in v1.

import (
	"fmt"
	"strings"
	"unicode"
)

// mergeRootToMain finishes a root land_to_main by merging the root's integration
// branch into the default branch (--no-ff) and deleting it (ADR 0058). It is a
// no-op for nodes without an open integration branch (ordinary land_to_main) or
// outside a git work tree. Called from finishLanding after the export commit.
func (s *Server) mergeRootToMain(nodeID string) error {
	if s.repo == nil {
		return nil
	}
	ib, err := s.db.GetIntegrationBranch(nodeID)
	if err != nil || ib == nil || ib.Status != "open" {
		return err
	}
	target := s.repo.DefaultBranch()
	if target == "" {
		return fmt.Errorf("cannot determine default branch to land %s into", nodeID)
	}
	if target == ib.Branch {
		// The integration target is already the default branch; nothing to merge.
		return s.db.CloseIntegrationBranch(nodeID)
	}
	if err := s.repo.Checkout(target); err != nil {
		return err
	}
	if err := s.repo.MergeNoFF(ib.Branch, fmt.Sprintf("Land %s into %s", nodeID, target)); err != nil {
		// A conflicted merge must not leave the work tree mid-conflict: abort to
		// restore the pre-merge state, then surface the conflict (ADR 0058). The
		// integration branch stays open so the human can resolve and retry.
		if abErr := s.repo.MergeAbort(); abErr != nil {
			return fmt.Errorf("merge of %s into %s failed (%v) and the abort also failed (%v); resolve the git state manually", ib.Branch, target, err, abErr)
		}
		return fmt.Errorf("merge of %s into %s failed and was aborted; resolve conflicts and retry land_to_main: %w", ib.Branch, target, err)
	}
	if err := s.repo.DeleteBranch(ib.Branch); err != nil {
		return err
	}
	return s.db.CloseIntegrationBranch(nodeID)
}

// ensureIntegrationBranch records (creating if needed) the integration target for
// a root node. It is idempotent and a no-op outside a git work tree.
func (s *Server) ensureIntegrationBranch(nodeID string) error {
	if s.repo == nil {
		return nil
	}
	if existing, err := s.db.GetIntegrationBranch(nodeID); err != nil || existing != nil {
		return err
	}
	cur, err := s.repo.CurrentBranch()
	if err != nil {
		return err
	}
	// Refuse to set up an integration target from a detached HEAD: we can neither
	// adopt it as a branch nor safely assume the operator's intent (they may be
	// mid-rebase or on a bare checkout). The human should check out a branch first.
	if cur == "HEAD" {
		return fmt.Errorf("cannot set up an integration branch for %s from a detached HEAD; check out a branch first", nodeID)
	}
	branch := cur
	if isDefaultBranch(cur) {
		// Don't land root work directly on the default branch: start a dedicated
		// integration branch from HEAD. `checkout -b` carries any in-flight work onto
		// the new branch (the operator's WIP becomes part of the feature), so no
		// dirty-tree guard here — and .groundwork/ is intentionally always
		// uncommitted, which would make such a guard misfire on every approval.
		branch = integrationBranchName(nodeID, s.ticketTitle(nodeID))
		if err := s.repo.CreateAndCheckout(branch); err != nil {
			return err
		}
	}
	base, err := s.repo.HeadCommit()
	if err != nil {
		return err
	}
	return s.db.RecordIntegrationBranch(nodeID, branch, base)
}

func (s *Server) ticketTitle(nodeID string) string {
	if t, err := s.db.GetTicket(nodeID); err == nil {
		return t.Title
	}
	return ""
}

func isDefaultBranch(b string) bool {
	switch b {
	case "", "main", "master":
		return true
	}
	return false
}

// integrationBranchName builds gw/root/<id>-<slug> (ADR 0058).
func integrationBranchName(nodeID, title string) string {
	name := "gw/root/" + nodeID
	if slug := slugify(title); slug != "" {
		name += "-" + slug
	}
	return name
}

// slugify lowercases title and collapses non-alphanumerics into single hyphens.
func slugify(title string) string {
	var b strings.Builder
	prevHyphen := false
	for _, r := range strings.ToLower(title) {
		switch {
		case unicode.IsLetter(r) || unicode.IsDigit(r):
			b.WriteRune(r)
			prevHyphen = false
		case !prevHyphen && b.Len() > 0:
			b.WriteByte('-')
			prevHyphen = true
		}
	}
	return strings.Trim(b.String(), "-")
}
