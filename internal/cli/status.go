package cli

import (
	"fmt"
	"sort"

	"groundwork/internal/approval"
	"groundwork/internal/store/sqlite"
	"groundwork/internal/ticket"
)

func newStatusCmd() *Command {
	return &Command{Name: "status", Usage: "Show a work-tree status summary", Run: runStatus}
}

func newBoardCmd() *Command {
	return &Command{Name: "board", Usage: "Show tickets grouped by status", Run: runBoard}
}

// indexByParent indexes tickets by id and groups them by parent id (the empty
// string key holds roots). Each child slice is sorted by id.
func indexByParent(all []*ticket.Ticket) (byID map[string]*ticket.Ticket, children map[string][]*ticket.Ticket) {
	byID = make(map[string]*ticket.Ticket, len(all))
	children = make(map[string][]*ticket.Ticket)
	for _, t := range all {
		byID[t.ID] = t
		children[t.ParentID] = append(children[t.ParentID], t)
	}
	for k := range children {
		sort.Slice(children[k], func(i, j int) bool { return children[k][i].ID < children[k][j].ID })
	}
	return byID, children
}

// rollupTree builds the child map and returns a function computing the derived
// rollup for any node id (memoized).
func rollupTree(all []*ticket.Ticket) (byID map[string]*ticket.Ticket, children map[string][]*ticket.Ticket, rollup func(string) ticket.Rollup) {
	byID, children = indexByParent(all)
	memo := map[string]ticket.Rollup{}
	var compute func(id string) ticket.Rollup
	compute = func(id string) ticket.Rollup {
		if r, ok := memo[id]; ok {
			return r
		}
		t, ok := byID[id]
		if !ok {
			return ticket.Rollup{}
		}
		// Provisional entry breaks any parent cycle before recursing.
		memo[id] = ticket.Rollup{}
		var childRollups []ticket.Rollup
		for _, c := range children[id] {
			childRollups = append(childRollups, compute(c.ID))
		}
		r := ticket.ComputeRollup(t.Status, childRollups)
		memo[id] = r
		return r
	}
	return byID, children, compute
}

func runStatus(ctx *Context, args []string) error {
	fs := ctx.NewFlagSet("gw status")
	if _, err := parseFlags(fs, args); err != nil {
		return err
	}

	p, db, err := ctx.openStore()
	if err != nil {
		return err
	}
	defer db.Close()

	all, err := db.ListTickets()
	if err != nil {
		return &Error{Code: "list_failed", Message: err.Error()}
	}

	counts := map[ticket.Status]int{}
	for _, t := range all {
		counts[t.Status]++
	}
	eligible, blocked, pending, err := readinessCounts(db, all)
	if err != nil {
		return &Error{Code: "store_error", Message: err.Error()}
	}
	_, children, rollup := rollupTree(all)

	type rootView struct {
		ID         string `json:"id"`
		Title      string `json:"title"`
		Status     string `json:"status"`
		HasBlocked bool   `json:"has_blocked"`
		HasActive  bool   `json:"has_active"`
	}
	var roots []rootView
	for _, t := range children[""] {
		r := rollup(t.ID)
		roots = append(roots, rootView{t.ID, t.Title, string(r.Status), r.HasBlocked, r.HasActive})
	}

	if ctx.JSON {
		countsOut := map[string]int{}
		for s, n := range counts {
			countsOut[string(s)] = n
		}
		return ctx.PrintJSON(map[string]any{
			"root":              p.Root,
			"total":             len(all),
			"counts":            countsOut,
			"eligible":          eligible,
			"blocked":           blocked,
			"pending_approvals": pending,
			"roots":             roots,
		})
	}

	fmt.Fprintf(ctx.Stdout, "Groundwork: %s\n", p.Root)
	fmt.Fprintf(ctx.Stdout, "Nodes: %d total\n", len(all))
	for _, s := range ticket.AllStatuses {
		if counts[s] > 0 {
			fmt.Fprintf(ctx.Stdout, "  %-12s %d\n", string(s)+":", counts[s])
		}
	}
	fmt.Fprintf(ctx.Stdout, "\nReady: %d   Blocked: %d   Pending approvals: %d\n", eligible, blocked, pending)
	if len(roots) > 0 {
		fmt.Fprintln(ctx.Stdout, "\nRoots (derived state):")
		for _, r := range roots {
			fmt.Fprintf(ctx.Stdout, "  %s  [%s]%s  %s\n", r.ID, r.Status, flags(r.HasBlocked, r.HasActive), r.Title)
		}
	}
	return nil
}

func runBoard(ctx *Context, args []string) error {
	fs := ctx.NewFlagSet("gw board")
	if _, err := parseFlags(fs, args); err != nil {
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

	groups := map[ticket.Status][]*ticket.Ticket{}
	for _, t := range all {
		groups[t.Status] = append(groups[t.Status], t)
	}
	for s := range groups {
		sort.Slice(groups[s], func(i, j int) bool { return groups[s][i].ID < groups[s][j].ID })
	}

	if ctx.JSON {
		out := map[string][]*ticket.Ticket{}
		for s, ts := range groups {
			out[string(s)] = ts
		}
		return ctx.PrintJSON(out)
	}

	for _, s := range ticket.AllStatuses {
		ts := groups[s]
		if len(ts) == 0 {
			continue
		}
		fmt.Fprintf(ctx.Stdout, "%s (%d)\n", s, len(ts))
		for _, t := range ts {
			fmt.Fprintf(ctx.Stdout, "  %s  %s\n", t.ID, t.Title)
		}
	}
	return nil
}

// readinessCounts summarizes what needs attention (ADR 0041): eligible (todo +
// dependencies satisfied), blocked (todo + a dependency unmet), and pending
// approvals — the "what can I do / what needs me" answer for gw status.
func readinessCounts(db *sqlite.DB, all []*ticket.Ticket) (eligible, blocked, pending int, err error) {
	for _, t := range all {
		if t.Status != ticket.StatusTodo {
			continue
		}
		ok, derr := db.DependenciesSatisfied(t.ID)
		if derr != nil {
			return 0, 0, 0, derr
		}
		if ok {
			eligible++
		} else {
			blocked++
		}
	}
	appr, aerr := db.ListApprovals(string(approval.StatusPending))
	if aerr != nil {
		return 0, 0, 0, aerr
	}
	return eligible, blocked, len(appr), nil
}

func flags(hasBlocked, hasActive bool) string {
	out := ""
	if hasBlocked {
		out += " (blocked)"
	}
	if hasActive {
		out += " (active)"
	}
	return out
}
