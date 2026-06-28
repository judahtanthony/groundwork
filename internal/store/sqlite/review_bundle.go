package sqlite

import (
	"groundwork/internal/completion"
	"groundwork/internal/ticket"
)

// ChildReview is one leaf's evidence in a review bundle (ADR 0057).
type ChildReview struct {
	NodeID      string              `json:"node_id"`
	Title       string              `json:"title"`
	Status      string              `json:"status"`
	Summary     *completion.Summary `json:"summary,omitempty"`
	Validations []*ValidationResult `json:"validations,omitempty"`
	Exceptions  []string            `json:"exceptions,omitempty"` // pending exception approval ids
}

// ReviewBundle is the deterministic feature-level review evidence for a
// composite/root node (ADR 0057): per-leaf summaries, validation, and exceptions,
// plus unresolved exceptions and a landing recommendation.
type ReviewBundle struct {
	NodeID               string        `json:"node_id"`
	Title                string        `json:"title"`
	Children             []ChildReview `json:"children"`
	UnresolvedExceptions []string      `json:"unresolved_exceptions,omitempty"`
	Recommendation       string        `json:"recommendation"` // land | hold | rework
}

// ReviewBundle assembles the review evidence for a node's subtree from existing
// records — read-only, no new authority (ADR 0057). The recommendation is hold
// when exceptions are unresolved, rework when a validation failed, else land.
func (db *DB) ReviewBundle(nodeID string) (*ReviewBundle, error) {
	root, err := db.GetTicket(nodeID)
	if err != nil {
		return nil, err
	}

	// Collect leaf descendants and the full subtree id set. The root node is part
	// of its own subtree, so an exception raised against the root itself is counted
	// in unresolved exceptions and the recommendation (L2/ADR 0057).
	subtree := map[string]bool{nodeID: true}
	var leaves []*ticket.Ticket
	var walk func(id string) error
	walk = func(id string) error {
		kids, err := db.ListChildren(id)
		if err != nil {
			return err
		}
		for _, k := range kids {
			subtree[k.ID] = true
			if k.NodeType == ticket.NodeLeaf {
				leaves = append(leaves, k)
			}
			if err := walk(k.ID); err != nil {
				return err
			}
		}
		return nil
	}
	if err := walk(nodeID); err != nil {
		return nil, err
	}

	// Pending exception approvals across the subtree, grouped by node.
	pending, err := db.ListApprovals("pending")
	if err != nil {
		return nil, err
	}
	exceptionsByNode := map[string][]string{}
	var unresolved []string
	for _, a := range pending {
		if a.Type == "exception" && subtree[a.TicketID] {
			exceptionsByNode[a.TicketID] = append(exceptionsByNode[a.TicketID], a.ID)
			unresolved = append(unresolved, a.ID)
		}
	}

	b := &ReviewBundle{NodeID: root.ID, Title: root.Title}
	anyFail := false
	for _, leaf := range leaves {
		summary, err := db.GetCompletionSummary(leaf.ID)
		if err != nil {
			return nil, err
		}
		vs, err := db.ListValidationsForTicket(leaf.ID)
		if err != nil {
			return nil, err
		}
		for _, v := range vs {
			if v.Status == ValidationFail {
				anyFail = true
			}
		}
		b.Children = append(b.Children, ChildReview{
			NodeID: leaf.ID, Title: leaf.Title, Status: string(leaf.Status),
			Summary: summary, Validations: vs, Exceptions: exceptionsByNode[leaf.ID],
		})
	}
	b.UnresolvedExceptions = unresolved
	switch {
	case len(unresolved) > 0:
		b.Recommendation = "hold"
	case anyFail:
		b.Recommendation = "rework"
	default:
		b.Recommendation = "land"
	}
	return b, nil
}
