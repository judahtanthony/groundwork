// Package contextbrief assembles the bounded, node-specific context brief that
// an agent receives at claim time (the read side of canon, ADR 0013): ancestor
// spine, parent contract, direct dependencies, relevant SOPs, and open
// escalations. It is the graph-assembly half of `gw context`; rendering lives in
// the CLI. Kept separate from internal/cli so the Phase 2 run supervisor can
// build briefs without importing the CLI.
package contextbrief

import (
	"os"
	"sort"

	"groundwork/internal/completion"
	"groundwork/internal/config"
	"groundwork/internal/decision"
	"groundwork/internal/store/sqlite"
	"groundwork/internal/ticket"
)

// Node is a compact reference to a work node within a brief.
type Node struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Status   string `json:"status"`
	NodeType string `json:"node_type,omitempty"`
}

// Brief is the bounded context assembled from canon via the SQLite graph.
type Brief struct {
	Node            Node     `json:"node"`
	Acceptance      []string `json:"acceptance"`
	AncestorSpine   []Node   `json:"ancestor_spine"`
	ParentContract  string   `json:"parent_contract,omitempty"`
	Dependencies    []Node   `json:"dependencies"`
	SOPs            []string `json:"sops"`
	OpenEscalations []string `json:"open_escalations"`
	Siblings        []Node   `json:"siblings,omitempty"`

	// Durable operational memory (ADR 0051/0047): the brief prefers these summaries
	// and canon over raw run transcripts.
	PendingBlockers    []decision.Record   `json:"pending_blockers,omitempty"`
	RecentDecisions    []decision.Record   `json:"recent_decisions,omitempty"`
	ChangedFiles       []string            `json:"changed_files,omitempty"`
	CompletionSummary  *completion.Summary `json:"completion_summary,omitempty"`
	SummaryStale       bool                `json:"summary_stale,omitempty"`
	SummaryStaleReason string              `json:"summary_stale_reason,omitempty"`
	// SummaryMissing flags a node at review/done with no completion summary — a
	// rework/recovery signal (ADR 0047).
	SummaryMissing bool `json:"summary_missing,omitempty"`
}

// recentDecisionLimit caps the resolved decision history carried in a brief, so
// the context stays bounded (newest kept).
const recentDecisionLimit = 5

// Build assembles the context brief for id. Escalations are always empty in
// Phase 1 (escalation events are Phase 2); the field is present for stability.
func Build(db *sqlite.DB, p *config.Project, id string, includeSiblings bool) (*Brief, error) {
	node, err := db.GetTicket(id)
	if err != nil {
		return nil, err
	}
	b := &Brief{
		Node:            toNode(node),
		Acceptance:      nonNilStrings(node.Acceptance),
		AncestorSpine:   []Node{},
		Dependencies:    []Node{},
		SOPs:            []string{},
		OpenEscalations: []string{},
	}

	ancestors, err := db.Ancestors(id)
	if err != nil {
		return nil, err
	}
	for _, a := range ancestors {
		b.AncestorSpine = append(b.AncestorSpine, toNode(a))
	}

	// Parent contract: the immediate parent's recorded contract, if any.
	if node.ParentID != "" {
		parent, err := db.GetTicket(node.ParentID)
		if err != nil {
			return nil, err
		}
		if parent.Contract != "" && parent.Contract != "{}" {
			b.ParentContract = parent.Contract
		}
	}

	depIDs, err := db.DependencyIDs(id)
	if err != nil {
		return nil, err
	}
	for _, depID := range depIDs {
		dep, err := db.GetTicket(depID)
		if err != nil {
			return nil, err
		}
		b.Dependencies = append(b.Dependencies, toNode(dep))
	}

	b.SOPs = relevantSOPs(p, node)

	// Durable decision records: pending blockers explain why the node waits; the
	// most recent resolved decisions are durable rationale (ADR 0051).
	recs, err := db.ListDecisions(id)
	if err != nil {
		return nil, err
	}
	var resolved []decision.Record
	for _, r := range recs {
		if r.Status == decision.StatusPending {
			b.PendingBlockers = append(b.PendingBlockers, r)
		} else {
			resolved = append(resolved, r)
		}
	}
	if n := len(resolved); n > recentDecisionLimit {
		resolved = resolved[n-recentDecisionLimit:]
	}
	b.RecentDecisions = resolved

	// Captured diff and completion summary, with a staleness/missing signal that
	// surfaces as a rework/recovery cue (ADR 0047).
	if files, ferr := db.ChangedFilesForNode(id); ferr == nil {
		b.ChangedFiles = files
	}
	if sum, ok, serr := completion.Read(p.TicketsDir(), id); serr == nil && ok {
		b.CompletionSummary = sum
		stale, reason := completion.Stale(sum, b.ChangedFiles, string(node.Status))
		b.SummaryStale, b.SummaryStaleReason = stale, reason
	} else if node.Status == ticket.StatusReview || node.Status == ticket.StatusDone {
		b.SummaryMissing = true
	}

	if includeSiblings && node.ParentID != "" {
		sibs, err := db.ListChildren(node.ParentID)
		if err != nil {
			return nil, err
		}
		b.Siblings = []Node{}
		for _, s := range sibs {
			if s.ID == id {
				continue
			}
			b.Siblings = append(b.Siblings, toNode(s))
		}
	}

	return b, nil
}

// relevantSOPs returns SOP directories under .groundwork/sops whose name matches
// the node's work_type (the primary key for SOPs, ADR 0023), falling back to its
// advisory kind or labels. Returns relative paths.
func relevantSOPs(p *config.Project, node *ticket.Ticket) []string {
	entries, err := os.ReadDir(p.SopsDir())
	if err != nil {
		return []string{}
	}
	want := map[string]bool{node.Kind: true}
	if node.WorkType != "" {
		want[node.WorkType] = true
	}
	for _, l := range node.Labels {
		want[l] = true
	}
	out := []string{}
	for _, e := range entries {
		if e.IsDir() && want[e.Name()] {
			out = append(out, "sops/"+e.Name()+"/")
		}
	}
	sort.Strings(out)
	return out
}

func toNode(t *ticket.Ticket) Node {
	return Node{ID: t.ID, Title: t.Title, Status: string(t.Status), NodeType: string(t.NodeType)}
}

func nonNilStrings(in []string) []string {
	if in == nil {
		return []string{}
	}
	return in
}
