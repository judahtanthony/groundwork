package server

// Root integration-branch lifecycle (ADR 0058). On envelope approval the root
// gets a recorded integration target: in v1's single working tree that is the
// current feature branch (the workflow dogfooded by hand), or a freshly created
// gw/root/<id>-<slug> branch when starting from the default branch. Children land
// to this branch (land_to_parent); the gated root land_to_main merges and cleans
// it up. No worktrees in v1.

import (
	"strings"
	"unicode"
)

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
	branch := cur
	if isDefaultBranch(cur) {
		// Don't land root work directly on the default branch: start a dedicated
		// integration branch from HEAD.
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
	case "", "main", "master", "HEAD":
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
