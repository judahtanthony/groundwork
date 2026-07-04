package server

// land_to_parent (ADR 0058): a child commits its work to its root's integration
// target, distinct from the human-gated land_to_main (which merges that branch to
// main). In v1's single working tree this generalizes ADR 0034's commit path,
// retargeted from main to the integration branch.

import (
	"errors"
	"fmt"
	"net/http"

	"groundwork/internal/store/sqlite"
	"groundwork/internal/worktree"
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
	// Diff-fed envelope enforcement (ADR 0056): the runtime supplied the child's
	// changed-file set, so enforce file scope and escalation triggers before
	// committing. A breach raises a human exception and blocks the landing.
	if _, err := s.enforceEnvelopeOnDiff(childID, "land_to_parent"); err != nil {
		return nil, err
	}
	// Landing mutates the single shared main working tree (checkout + merge --squash
	// + commit). Serialize all working-tree mutations so concurrent landings cannot
	// race on one index (review finding #5).
	s.repoMu.Lock()
	defer s.repoMu.Unlock()

	if s.repo != nil {
		if cur, _ := s.repo.CurrentBranch(); cur != ib.Branch {
			if err := s.repo.Checkout(ib.Branch); err != nil {
				return nil, err
			}
		}
	}
	// Stage the run branch into the integration branch FIRST (the conflict-prone
	// step), so a squash conflict leaves the child untouched in review — re-landable
	// — rather than recorded done-but-uncommitted (review finding #6). Without a run
	// branch (manual/human work in the single tree) this is a no-op and the
	// working-tree commit path is used unchanged (ADR 0058).
	runID, err := s.squashRunBranch(childID)
	if err != nil {
		return nil, err
	}
	// Only after the work is staged do we mark the child done, so commitLanding's
	// regenerated export reflects done and is committed together with the squashed
	// change as one landing commit. Use the land_to_parent status primitive (not a
	// single-hop transition): the scheduler leaves a completed run's node in review,
	// and review -> done is not a legal single hop (it must pass through landing).
	if err := s.db.LandToParentDone(childID, ownerActor); err != nil {
		return nil, err
	}
	if err := s.commitLanding(childID); err != nil {
		return nil, err
	}
	// Retain the WIP checkpoint chain under the run ref, then tear down the run
	// branch and worktree now that the work has landed (ADR 0015/0059).
	if runID != "" {
		s.cleanupLandedRun(childID, runID)
	}
	return ib, nil
}

// squashRunBranch squash-merges the node's latest run branch into the currently
// checked-out integration branch's index, returning the run id (or "" when the
// node has no run branch to squash). A squash conflict is surfaced after resetting
// the index so the tree is not left mid-conflict.
func (s *Server) squashRunBranch(nodeID string) (string, error) {
	if s.repo == nil {
		return "", nil
	}
	runID, err := s.db.LatestRunIDForNode(nodeID)
	if err != nil || runID == "" {
		return "", err
	}
	branch := worktree.RunBranch(runID)
	if !s.repo.BranchExists(branch) {
		return "", nil
	}
	if err := s.repo.MergeSquash(branch); err != nil {
		_ = s.repo.ResetHard("HEAD")
		return "", fmt.Errorf("squash %s into integration branch (resolve conflicts manually): %w", branch, err)
	}
	return runID, nil
}

// cleanupLandedRun retains the run's WIP chain under its ref and removes the run
// branch and worktree after a successful landing (best-effort; failures surface
// as audit events rather than failing the completed landing).
func (s *Server) cleanupLandedRun(nodeID, runID string) {
	mgr := worktree.NewManager(s.repo, s.proj.WorktreesDir())
	if err := mgr.Retain(runID); err != nil {
		_ = s.db.AppendAuditEvent(ownerActor, "run.cleanup_error", "run", runID, map[string]any{"op": "retain", "error": err.Error()})
	}
	if err := mgr.Teardown(runID, true); err != nil {
		_ = s.db.AppendAuditEvent(ownerActor, "run.cleanup_error", "run", runID, map[string]any{"op": "teardown", "error": err.Error()})
	}
}

// landRouteFor reports how `gw ticket land` should land a node (ADR 0058): a node
// whose nearest integration target is an ANCESTOR (not itself) is a child that
// lands to its root's integration branch ("parent"); anything else — a root that
// owns its integration branch, or a node with no integration chain — lands to main
// ("main"). runBranch reports whether the node has a live gw/run/<id> branch whose
// work would be orphaned by a wrong land-to-main; the CLI uses route to skip
// main-tree staging and route the landing correctly.
func (s *Server) landRouteFor(nodeID string) (route, branch string, runBranch bool, err error) {
	if _, err := s.db.GetTicket(nodeID); err != nil {
		return "", "", false, err
	}
	ib, err := s.integrationTargetFor(nodeID)
	if err != nil {
		return "", "", false, err
	}
	if ib == nil || ib.NodeID == nodeID {
		return "main", "", false, nil
	}
	if s.repo != nil {
		if runID, rerr := s.db.LatestRunIDForNode(nodeID); rerr == nil && runID != "" {
			runBranch = s.repo.BranchExists(worktree.RunBranch(runID))
		}
	}
	return "parent", ib.Branch, runBranch, nil
}

// handleTicketLandRoute reports the landing route for a node so the CLI can pick
// land_to_parent over a main-tree commit for run-backed children (ADR 0058).
func (s *Server) handleTicketLandRoute(w http.ResponseWriter, r *http.Request) {
	route, branch, runBranch, err := s.landRouteFor(r.PathValue("id"))
	if err != nil {
		s.writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"route":              route,
		"integration_branch": branch,
		"run_branch":         runBranch,
	})
}

// handleTicketLandToParent lands a child to its root integration target.
func (s *Server) handleTicketLandToParent(w http.ResponseWriter, r *http.Request) {
	ib, err := s.LandToParent(r.PathValue("id"))
	if err != nil {
		switch {
		case errors.Is(err, errNoIntegrationTarget):
			writeError(w, http.StatusConflict, "no_integration_target", err.Error())
		case errors.Is(err, ErrEnvelopeEscalation):
			writeError(w, http.StatusConflict, "envelope_escalation", err.Error())
		default:
			s.writeMutationError(w, err)
		}
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"landed_to": ib.Branch, "node_id": ib.NodeID})
}
