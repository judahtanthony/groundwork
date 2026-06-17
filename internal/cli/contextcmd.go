package cli

import (
	"fmt"
	"os"
	"sort"

	"groundwork/internal/config"
	"groundwork/internal/store/sqlite"
	"groundwork/internal/ticket"
)

// newContextCmd is the top-level `gw context <id>`.
func newContextCmd() *Command {
	return &Command{Name: "context", Usage: "Show the bounded context brief for a node", Args: "<id>", Run: runContext}
}

// newTicketContextCmd is `gw ticket context <id>`.
func newTicketContextCmd() *Command {
	return &Command{Name: "context", Usage: "Show the bounded context brief for a node", Args: "<id>", Run: runContext}
}

// briefNode is a compact reference to a node within a brief.
type briefNode struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Status   string `json:"status"`
	NodeType string `json:"node_type,omitempty"`
}

// Brief is the bounded, node-specific context assembled from canon via the
// SQLite graph (ADR 0013). It is what an agent receives at claim time.
type Brief struct {
	Node            briefNode   `json:"node"`
	AncestorSpine   []briefNode `json:"ancestor_spine"`
	ParentContract  string      `json:"parent_contract,omitempty"`
	Dependencies    []briefNode `json:"dependencies"`
	SOPs            []string    `json:"sops"`
	OpenEscalations []string    `json:"open_escalations"`
	Siblings        []briefNode `json:"siblings,omitempty"`
}

func runContext(ctx *Context, args []string) error {
	fs := ctx.NewFlagSet("gw context")
	var siblings bool
	fs.BoolVar(&siblings, "siblings", false, "include sibling nodes (off by default)")
	pos, err := parseFlags(fs, args)
	if err != nil {
		return err
	}
	if len(pos) < 1 {
		return &Error{Code: "invalid_args", Message: "usage: gw context <id> [--siblings]"}
	}

	p, db, err := ctx.openStore()
	if err != nil {
		return err
	}
	defer db.Close()

	brief, err := buildBrief(db, p, pos[0], siblings)
	if err != nil {
		return ticketError(err, pos[0])
	}

	if ctx.JSON {
		return ctx.PrintJSON(brief)
	}
	renderBrief(ctx, brief)
	return nil
}

// buildBrief assembles the context brief for id. Escalations are always empty in
// Phase 1 (escalation events are Phase 2); the section is present for stability.
func buildBrief(db *sqlite.DB, p *config.Project, id string, includeSiblings bool) (*Brief, error) {
	node, err := db.GetTicket(id)
	if err != nil {
		return nil, err
	}
	b := &Brief{
		Node:            toBriefNode(node),
		AncestorSpine:   []briefNode{},
		Dependencies:    []briefNode{},
		SOPs:            []string{},
		OpenEscalations: []string{},
	}

	ancestors, err := db.Ancestors(id)
	if err != nil {
		return nil, err
	}
	for _, a := range ancestors {
		b.AncestorSpine = append(b.AncestorSpine, toBriefNode(a))
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
		b.Dependencies = append(b.Dependencies, toBriefNode(dep))
	}

	b.SOPs = relevantSOPs(p, node)

	if includeSiblings && node.ParentID != "" {
		sibs, err := db.ListChildren(node.ParentID)
		if err != nil {
			return nil, err
		}
		b.Siblings = []briefNode{}
		for _, s := range sibs {
			if s.ID == id {
				continue
			}
			b.Siblings = append(b.Siblings, toBriefNode(s))
		}
	}

	return b, nil
}

// relevantSOPs returns SOP directories under .groundwork/sops whose name matches
// the node's advisory kind or one of its labels. Returns relative paths.
func relevantSOPs(p *config.Project, node *ticket.Ticket) []string {
	entries, err := os.ReadDir(p.SopsDir())
	if err != nil {
		return []string{}
	}
	want := map[string]bool{node.Kind: true}
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

func toBriefNode(t *ticket.Ticket) briefNode {
	return briefNode{ID: t.ID, Title: t.Title, Status: string(t.Status), NodeType: string(t.NodeType)}
}

func renderBrief(ctx *Context, b *Brief) {
	w := ctx.Stdout
	fmt.Fprintf(w, "Context for %s  %s\n", b.Node.ID, b.Node.Title)
	fmt.Fprintf(w, "  status: %s  type: %s\n\n", b.Node.Status, orDash(b.Node.NodeType, "untriaged"))

	fmt.Fprintln(w, "Ancestor spine:")
	if len(b.AncestorSpine) == 0 {
		fmt.Fprintln(w, "  (root node)")
	} else {
		for _, a := range b.AncestorSpine {
			fmt.Fprintf(w, "  %s  %s\n", a.ID, a.Title)
		}
	}

	fmt.Fprintln(w, "\nParent contract:")
	if b.ParentContract == "" {
		fmt.Fprintln(w, "  (none)")
	} else {
		fmt.Fprintf(w, "  %s\n", b.ParentContract)
	}

	fmt.Fprintln(w, "\nDependencies:")
	if len(b.Dependencies) == 0 {
		fmt.Fprintln(w, "  (none)")
	} else {
		for _, d := range b.Dependencies {
			fmt.Fprintf(w, "  %s  %s  [%s]\n", d.ID, d.Title, d.Status)
		}
	}

	fmt.Fprintln(w, "\nSOPs:")
	if len(b.SOPs) == 0 {
		fmt.Fprintln(w, "  (none)")
	} else {
		for _, s := range b.SOPs {
			fmt.Fprintf(w, "  %s\n", s)
		}
	}

	fmt.Fprintln(w, "\nOpen escalations:")
	if len(b.OpenEscalations) == 0 {
		fmt.Fprintln(w, "  (none)")
	} else {
		for _, e := range b.OpenEscalations {
			fmt.Fprintf(w, "  %s\n", e)
		}
	}

	if b.Siblings != nil {
		fmt.Fprintln(w, "\nSiblings:")
		if len(b.Siblings) == 0 {
			fmt.Fprintln(w, "  (none)")
		} else {
			for _, s := range b.Siblings {
				fmt.Fprintf(w, "  %s  %s  [%s]\n", s.ID, s.Title, s.Status)
			}
		}
	}
}
