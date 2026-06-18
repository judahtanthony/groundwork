// Package ticket holds the work-node domain model. A "ticket" is a uniform work
// node (ADR 0009): kind is an advisory label, node_type is the structural fact
// (leaf | composite) set at triage.
package ticket

// NodeType is the structural classification set at triage.
type NodeType string

const (
	// NodeUnset means the node has not been triaged yet.
	NodeUnset NodeType = ""
	// NodeLeaf is one verifiable change, dispatched to an executing agent.
	NodeLeaf NodeType = "leaf"
	// NodeComposite decomposes into children.
	NodeComposite NodeType = "composite"
)

// Valid reports whether n is a recognized node type (unset is allowed).
func (n NodeType) Valid() bool {
	switch n {
	case NodeUnset, NodeLeaf, NodeComposite:
		return true
	}
	return false
}

// Ticket is a single work node. Nullable database columns are represented by
// the empty string (text) or nil pointer (integers).
type Ticket struct {
	ID       string   `json:"id"`
	ParentID string   `json:"parent_id,omitempty"`
	Kind     string   `json:"kind"`
	NodeType NodeType `json:"node_type,omitempty"`
	// WorkType is organization-defined operational metadata (ADR 0023) used by
	// SOPs, policy, actor routing, and validation. It is not a status.
	WorkType    string `json:"work_type,omitempty"`
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
	// Contract holds the parent contract for composite nodes as canonical JSON
	// (default "{}"). It is opaque to Phase 1.
	Contract string `json:"contract,omitempty"`
	Status   Status `json:"status"`
	Assignee string `json:"assignee,omitempty"`
	// RequestedActor is an optional routing hint naming a preferred actor; policy
	// must still authorize the claim (ADR 0023). Distinct from Assignee, which is
	// a display-only ownership label.
	RequestedActor string   `json:"requested_actor,omitempty"`
	Priority       *int     `json:"priority,omitempty"`
	Labels         []string `json:"labels"`
	Acceptance     []string `json:"acceptance"`
	RiskScore      *int     `json:"risk_score,omitempty"`
	CreatedAt      string   `json:"created_at"`
	UpdatedAt      string   `json:"updated_at"`
}
