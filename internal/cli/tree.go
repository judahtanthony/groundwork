package cli

import (
	"fmt"
	"sort"

	"groundwork/internal/ticket"
)

func newTicketTreeCmd() *Command {
	return &Command{Name: "tree", Usage: "Show the work-tree hierarchy", Args: "[id]", Run: runTicketTree}
}

func runTicketTree(ctx *Context, args []string) error {
	fs := ctx.NewFlagSet("gw ticket tree")
	pos, err := parseFlags(fs, args)
	if err != nil {
		return err
	}

	_, db, err := ctx.openStore()
	if err != nil {
		return err
	}
	defer db.Close()

	all, err := db.ListTickets()
	if err != nil {
		return &Error{Code: "list_failed", Message: err.Error()}
	}

	byID := make(map[string]*ticket.Ticket, len(all))
	children := make(map[string][]*ticket.Ticket)
	for _, t := range all {
		byID[t.ID] = t
		children[t.ParentID] = append(children[t.ParentID], t)
	}
	for k := range children {
		sort.Slice(children[k], func(i, j int) bool { return children[k][i].ID < children[k][j].ID })
	}

	// Determine the roots to print: a given subtree, or all top-level nodes.
	var roots []*ticket.Ticket
	if len(pos) >= 1 {
		t, ok := byID[pos[0]]
		if !ok {
			return &Error{Code: "not_found", Message: fmt.Sprintf("ticket %q not found", pos[0])}
		}
		roots = []*ticket.Ticket{t}
	} else {
		roots = children[""] // nodes with no parent
	}

	if ctx.JSON {
		return ctx.PrintJSON(buildTreeNodes(roots, children))
	}

	if len(roots) == 0 {
		fmt.Fprintln(ctx.Stdout, "No tickets.")
		return nil
	}
	for _, r := range roots {
		printTree(ctx, r, children, 0)
	}
	return nil
}

// treeNode is the JSON shape for `gw ticket tree --json`.
type treeNode struct {
	ID       string     `json:"id"`
	Title    string     `json:"title"`
	Status   string     `json:"status"`
	NodeType string     `json:"node_type,omitempty"`
	Children []treeNode `json:"children"`
}

func buildTreeNodes(roots []*ticket.Ticket, children map[string][]*ticket.Ticket) []treeNode {
	out := []treeNode{}
	for _, r := range roots {
		n := treeNode{ID: r.ID, Title: r.Title, Status: string(r.Status), NodeType: string(r.NodeType)}
		n.Children = buildTreeNodes(children[r.ID], children)
		out = append(out, n)
	}
	return out
}

func printTree(ctx *Context, t *ticket.Ticket, children map[string][]*ticket.Ticket, depth int) {
	indent := ""
	for i := 0; i < depth; i++ {
		indent += "  "
	}
	fmt.Fprintf(ctx.Stdout, "%s%s  [%s] %s  %s\n",
		indent, t.ID, orDash(string(t.NodeType), "untriaged"), t.Status, t.Title)
	for _, c := range children[t.ID] {
		printTree(ctx, c, children, depth+1)
	}
}
