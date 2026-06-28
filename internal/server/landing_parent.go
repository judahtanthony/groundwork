package server

// land_to_parent (ADR 0058): a child commits its work to its root's integration
// target, distinct from the human-gated land_to_main (which merges that branch to
// main). In v1's single working tree this generalizes ADR 0034's commit path,
// retargeted from main to the integration branch.

import (
	"errors"
	"net/http"

	"groundwork/internal/store/sqlite"
	"groundwork/internal/ticket"
)

var errNoIntegrationTarget = errors.New("no integration target: approve a root envelope first")

// integrationTargetFor returns the integration branch nearest to nodeID (self,
// then closest ancestor), or (nil, nil) when none exists in the node's chain.
func (s *Server) integrationTargetFor(nodeID string) (*sqlite.IntegrationBranch, error) {
	return nearestInChain(s, nodeID, s.db.GetIntegrationBranch)
}

// LandToParent lands an in-progress child onto its root integration branch: mark
// it done and commit its export plus staged work to that branch (ADR 0058/0034).
// It does not merge to main and does not open the land_to_main gate.
func (s *Server) LandToParent(childID string) (*sqlite.IntegrationBranch, error) {
	if _, err := s.db.GetTicket(childID); err != nil {
		return nil, err
	}
	ib, err := s.integrationTargetFor(childID)
	if err != nil {
		return nil, err
	}
	if ib == nil {
		return nil, errNoIntegrationTarget
	}
	// Enforce the validation gate before committing to the integration branch, the
	// same "no failing results" bar land_to_main applies (ADR 0058): land_to_parent
	// is a lighter landing level, not an unguarded one. M2 supplies no required
	// checks, so this blocks children with a failing validation result.
	if _, err := s.db.CheckValidationGate(childID, nil, false); err != nil {
		return nil, err
	}
	if s.repo != nil {
		if cur, _ := s.repo.CurrentBranch(); cur != ib.Branch {
			if err := s.repo.Checkout(ib.Branch); err != nil {
				return nil, err
			}
		}
	}
	if err := s.db.TransitionTicket(childID, ticket.StatusDone, ownerActor); err != nil {
		return nil, err
	}
	if err := s.commitLanding(childID); err != nil {
		return nil, err
	}
	return ib, nil
}

// handleTicketLandToParent lands a child to its root integration target.
func (s *Server) handleTicketLandToParent(w http.ResponseWriter, r *http.Request) {
	ib, err := s.LandToParent(r.PathValue("id"))
	if err != nil {
		switch {
		case errors.Is(err, errNoIntegrationTarget):
			writeError(w, http.StatusConflict, "no_integration_target", err.Error())
		default:
			s.writeMutationError(w, err)
		}
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"landed_to": ib.Branch, "node_id": ib.NodeID})
}
