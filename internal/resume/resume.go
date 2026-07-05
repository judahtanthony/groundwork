// Package resume assembles a structured resume packet for a new run from durable
// state (ADR 0051): the ticket context, ancestor contract, acceptance criteria,
// dependency status, resolved/pending decision records, prior handoff summary,
// captured diff, and validation state — plus a recommended next action. Resume
// means "start a new run from durable context," not "continue the old session."
package resume

import (
	"groundwork/internal/decision"
	"groundwork/internal/store/sqlite"
)

// DepStatus is a dependency and its current status.
type DepStatus struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

// Validation is a validation result name + status.
type Validation struct {
	Name   string `json:"name"`
	Status string `json:"status"`
}

// Packet is the durable context handed to a resuming run (ADR 0051).
type Packet struct {
	TicketID          string            `json:"ticket_id"`
	Title             string            `json:"title"`
	WorkType          string            `json:"work_type"`
	Status            string            `json:"status"`
	Acceptance        []string          `json:"acceptance,omitempty"`
	AncestorContract  string            `json:"ancestor_contract,omitempty"`
	Dependencies      []DepStatus       `json:"dependencies,omitempty"`
	PendingBlockers   []decision.Record `json:"pending_blockers,omitempty"`
	ResolvedDecisions []decision.Record `json:"resolved_decisions,omitempty"`
	HandoffSummary    string            `json:"handoff_summary,omitempty"`
	ChangedFiles      []string          `json:"changed_files,omitempty"`
	Validations       []Validation      `json:"validations,omitempty"`
	NextAction        string            `json:"next_action"`
}

// Assemble builds the resume packet for nodeID from durable store state.
func Assemble(db *sqlite.DB, nodeID string) (*Packet, error) {
	t, err := db.GetTicket(nodeID)
	if err != nil {
		return nil, err
	}
	p := &Packet{
		TicketID: t.ID, Title: t.Title, WorkType: t.WorkType,
		Status: string(t.Status), Acceptance: t.Acceptance,
	}

	// Nearest ancestor contract (the forward channel the node implements).
	ancestors, err := db.Ancestors(nodeID)
	if err != nil {
		return nil, err
	}
	for _, a := range ancestors {
		if a.Contract != "" && a.Contract != "{}" {
			p.AncestorContract = a.Contract
			break
		}
	}

	depIDs, err := db.DependencyIDs(nodeID)
	if err != nil {
		return nil, err
	}
	for _, id := range depIDs {
		st := ""
		if dt, derr := db.GetTicket(id); derr == nil {
			st = string(dt.Status)
		}
		p.Dependencies = append(p.Dependencies, DepStatus{ID: id, Status: st})
	}

	// Decision records: split pending blockers from resolved history; the latest
	// pending blocker (or the latest record's handoff summary) is the handoff.
	recs, err := db.ListDecisions(nodeID)
	if err != nil {
		return nil, err
	}
	for _, r := range recs {
		if r.Status == decision.StatusPending {
			p.PendingBlockers = append(p.PendingBlockers, r)
			if r.HandoffSummary != "" {
				p.HandoffSummary = r.HandoffSummary
			}
		} else {
			p.ResolvedDecisions = append(p.ResolvedDecisions, r)
		}
	}
	if p.HandoffSummary == "" {
		for i := len(recs) - 1; i >= 0; i-- {
			if recs[i].HandoffSummary != "" {
				p.HandoffSummary = recs[i].HandoffSummary
				break
			}
		}
	}

	if files, ferr := db.ChangedFilesForNode(nodeID); ferr == nil {
		p.ChangedFiles = files
	}

	vals, err := db.ListValidationsForTicket(nodeID)
	if err != nil {
		return nil, err
	}
	for _, v := range vals {
		p.Validations = append(p.Validations, Validation{Name: v.Name, Status: string(v.Status)})
	}

	p.NextAction = nextAction(p)
	return p, nil
}

// nextAction recommends the resuming run's first step from durable signals.
func nextAction(p *Packet) string {
	if len(p.PendingBlockers) > 0 {
		if p.HandoffSummary != "" {
			return "resolve blocker: " + p.HandoffSummary
		}
		return "resolve the pending blocker before continuing"
	}
	for _, v := range p.Validations {
		if v.Status == string(sqlite.ValidationFail) {
			return "fix failing validation: " + v.Name
		}
	}
	if p.Status == "rework" {
		return "address rework feedback and re-submit for review"
	}
	return "continue implementation toward the acceptance criteria"
}
