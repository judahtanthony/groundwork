package cli

import (
	"errors"
	"fmt"
	"strings"

	"groundwork/internal/contextbrief"
	"groundwork/internal/store/sqlite"
	"groundwork/internal/ticket"
)

func newTicketCreateCmd() *Command {
	return &Command{Name: "create", Usage: "Create a work node", Run: runTicketCreate}
}

func runTicketCreate(ctx *Context, args []string) error {
	fs := ctx.NewFlagSet("gw ticket create")
	var (
		title, kind, parent, status, desc, assignee string
		workType, requestedActor                    string
		priority                                    float64
		labels, acceptance                          stringSlice
	)
	fs.StringVar(&title, "title", "", "node title (required)")
	fs.StringVar(&kind, "kind", "", "advisory kind label (default: ticket)")
	fs.StringVar(&parent, "parent", "", "parent node id")
	fs.StringVar(&status, "status", "", "initial status (default: backlog)")
	fs.StringVar(&desc, "description", "", "description")
	fs.StringVar(&assignee, "assignee", "", "assignee (display-only ownership label)")
	fs.StringVar(&workType, "work-type", "", "operational work type (e.g. technical_implementation)")
	fs.StringVar(&requestedActor, "requested-actor", "", "preferred actor id (routing hint, still policy-checked)")
	fs.Float64Var(&priority, "priority", 0, "value priority in [0,1]; higher runs first (ADR 0039)")
	fs.Var(&labels, "label", "label (repeatable)")
	fs.Var(&acceptance, "acceptance", "acceptance criterion (repeatable)")
	if err := fs.Parse(args); err != nil {
		return err
	}
	set := setFlags(fs)

	if title == "" {
		return &Error{Code: "invalid_args", Message: "--title is required"}
	}
	if status != "" && !ticket.Status(status).Valid() {
		return &Error{Code: "invalid_args", Message: fmt.Sprintf("invalid status %q", status)}
	}
	if set["priority"] {
		if err := validatePriority(priority); err != nil {
			return err
		}
	}

	t := &ticket.Ticket{
		Title:          title,
		Kind:           kind,
		ParentID:       parent,
		Status:         ticket.Status(status),
		Description:    desc,
		Assignee:       assignee,
		WorkType:       workType,
		RequestedActor: requestedActor,
		Labels:         labels,
		Acceptance:     acceptance,
	}
	if set["priority"] {
		t.Priority = &priority
	}

	store, closeStore, err := ctx.openTicketStore()
	if err != nil {
		return err
	}
	defer closeStore()

	if err := store.CreateTicket(t, ownerActor); err != nil {
		return &Error{Code: "create_failed", Message: err.Error()}
	}

	if ctx.JSON {
		return ctx.PrintJSON(t)
	}
	fmt.Fprintf(ctx.Stdout, "Created %s  %s\n", t.ID, t.Title)
	return nil
}

func newTicketShowCmd() *Command {
	return &Command{Name: "show", Usage: "Show a work node", Args: "<id>", Run: runTicketShow}
}

func runTicketShow(ctx *Context, args []string) error {
	fs := ctx.NewFlagSet("gw ticket show")
	pos, err := parseFlags(fs, args)
	if err != nil {
		return err
	}
	if len(pos) < 1 {
		return &Error{Code: "invalid_args", Message: "usage: gw ticket show <id>"}
	}

	_, db, err := ctx.openStore()
	if err != nil {
		return err
	}
	defer db.Close()

	t, err := db.GetTicket(pos[0])
	if err != nil {
		return ticketError(err, pos[0])
	}
	deps, err := db.DependencyIDs(t.ID)
	if err != nil {
		return &Error{Code: "store_error", Message: err.Error()}
	}
	if ctx.JSON {
		return ctx.PrintJSON(struct {
			*ticket.Ticket
			DependsOn []string `json:"depends_on"`
		}{t, deps})
	}
	renderTicket(ctx, t, deps)
	return nil
}

func newTicketListCmd() *Command {
	return &Command{Name: "list", Usage: "List work nodes", Run: runTicketList}
}

func runTicketList(ctx *Context, args []string) error {
	fs := ctx.NewFlagSet("gw ticket list")
	var status string
	var ready, blocked bool
	fs.StringVar(&status, "status", "", "filter by status")
	fs.BoolVar(&ready, "ready", false, "show only eligible nodes (todo + deps satisfied), value-ordered")
	fs.BoolVar(&blocked, "blocked", false, "show only todo nodes blocked by unsatisfied dependencies")
	if err := fs.Parse(args); err != nil {
		return err
	}
	modes := 0
	for _, on := range []bool{status != "", ready, blocked} {
		if on {
			modes++
		}
	}
	if modes > 1 {
		return &Error{Code: "invalid_args", Message: "--status, --ready, and --blocked are mutually exclusive"}
	}
	if status != "" && !ticket.Status(status).Valid() {
		return &Error{Code: "invalid_args", Message: fmt.Sprintf("invalid status %q", status)}
	}

	_, db, err := ctx.openStore()
	if err != nil {
		return err
	}
	defer db.Close()

	switch {
	case ready:
		return listReady(ctx, db)
	case blocked:
		return listBlocked(ctx, db)
	}

	all, err := db.ListTickets()
	if err != nil {
		return &Error{Code: "list_failed", Message: err.Error()}
	}
	var tickets []*ticket.Ticket
	for _, t := range all {
		if status != "" && string(t.Status) != status {
			continue
		}
		tickets = append(tickets, t)
	}

	if ctx.JSON {
		if tickets == nil {
			tickets = []*ticket.Ticket{}
		}
		return ctx.PrintJSON(tickets)
	}
	renderTicketList(ctx, tickets, "No tickets.")
	return nil
}

// listReady prints the eligible set (todo + dependencies satisfied) in ADR 0039
// value order — the human-facing read of the same surface the scheduler
// dispatches from (ADR 0041).
func listReady(ctx *Context, db *sqlite.DB) error {
	tickets, err := db.ListEligibleOrdered()
	if err != nil {
		return &Error{Code: "list_failed", Message: err.Error()}
	}
	if ctx.JSON {
		if tickets == nil {
			tickets = []*ticket.Ticket{}
		}
		return ctx.PrintJSON(tickets)
	}
	renderTicketList(ctx, tickets, "No ready nodes.")
	return nil
}

// blocker names a dependency that is keeping a node out of the eligible set.
type blocker struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

// blockedNode is a todo node plus the unsatisfied dependencies blocking it.
type blockedNode struct {
	*ticket.Ticket
	BlockedBy []blocker `json:"blocked_by"`
}

// listBlocked prints todo nodes that are not eligible because one or more
// dependencies are not yet done, annotated with the blocking deps (ADR 0041).
func listBlocked(ctx *Context, db *sqlite.DB) error {
	all, err := db.ListTickets()
	if err != nil {
		return &Error{Code: "list_failed", Message: err.Error()}
	}
	nodes := []blockedNode{}
	for _, t := range all {
		if t.Status != ticket.StatusTodo {
			continue
		}
		depIDs, err := db.DependencyIDs(t.ID)
		if err != nil {
			return &Error{Code: "store_error", Message: err.Error()}
		}
		var blockers []blocker
		for _, depID := range depIDs {
			dep, err := db.GetTicket(depID)
			if err != nil {
				return &Error{Code: "store_error", Message: err.Error()}
			}
			if !ticket.DependencyMet(dep.Status) {
				blockers = append(blockers, blocker{ID: dep.ID, Status: string(dep.Status)})
			}
		}
		if len(blockers) > 0 {
			nodes = append(nodes, blockedNode{Ticket: t, BlockedBy: blockers})
		}
	}

	if ctx.JSON {
		return ctx.PrintJSON(nodes)
	}
	if len(nodes) == 0 {
		fmt.Fprintln(ctx.Stdout, "No blocked nodes.")
		return nil
	}
	fmt.Fprintf(ctx.Stdout, "%-8s  %-9s  %s\n", "ID", "TYPE", "TITLE")
	for _, n := range nodes {
		fmt.Fprintf(ctx.Stdout, "%-8s  %-9s  %s\n",
			n.ID, orDash(string(n.NodeType), "-"), n.Title)
		parts := make([]string, len(n.BlockedBy))
		for i, b := range n.BlockedBy {
			parts[i] = fmt.Sprintf("%s (%s)", b.ID, b.Status)
		}
		fmt.Fprintf(ctx.Stdout, "%-8s  blocked by: %s\n", "", strings.Join(parts, ", "))
	}
	return nil
}

func newTicketEditCmd() *Command {
	return &Command{Name: "edit", Usage: "Edit a work node", Args: "<id>", Run: runTicketEdit}
}

func runTicketEdit(ctx *Context, args []string) error {
	fs := ctx.NewFlagSet("gw ticket edit")
	var (
		title, kind, desc, assignee string
		workType, requestedActor    string
		priority                    float64
		labels, acceptance          stringSlice
	)
	fs.StringVar(&title, "title", "", "new title")
	fs.StringVar(&kind, "kind", "", "new advisory kind")
	fs.StringVar(&desc, "description", "", "new description")
	fs.StringVar(&assignee, "assignee", "", "new assignee (display-only label)")
	fs.StringVar(&workType, "work-type", "", "new work type")
	fs.StringVar(&requestedActor, "requested-actor", "", "new requested actor (routing hint)")
	fs.Float64Var(&priority, "priority", 0, "new value priority in [0,1]; higher runs first")
	fs.Var(&labels, "label", "replace labels (repeatable)")
	fs.Var(&acceptance, "acceptance", "replace acceptance (repeatable)")
	pos, err := parseFlags(fs, args)
	if err != nil {
		return err
	}
	if len(pos) < 1 {
		return &Error{Code: "invalid_args", Message: "usage: gw ticket edit <id> [flags]"}
	}
	set := setFlags(fs)

	store, closeStore, err := ctx.openTicketStore()
	if err != nil {
		return err
	}
	defer closeStore()

	t, err := store.GetTicket(pos[0])
	if err != nil {
		return ticketError(err, pos[0])
	}

	if set["title"] {
		t.Title = title
	}
	if set["kind"] {
		t.Kind = kind
	}
	if set["description"] {
		t.Description = desc
	}
	if set["assignee"] {
		t.Assignee = assignee
	}
	if set["work-type"] {
		t.WorkType = workType
	}
	if set["requested-actor"] {
		t.RequestedActor = requestedActor
	}
	if set["priority"] {
		if err := validatePriority(priority); err != nil {
			return err
		}
		t.Priority = &priority
	}
	if set["label"] {
		t.Labels = labels
	}
	if set["acceptance"] {
		t.Acceptance = acceptance
	}

	if err := store.UpdateTicket(t, ownerActor); err != nil {
		return &Error{Code: "edit_failed", Message: err.Error()}
	}
	if ctx.JSON {
		return ctx.PrintJSON(t)
	}
	fmt.Fprintf(ctx.Stdout, "Updated %s\n", t.ID)
	return nil
}

func newTicketClaimCmd() *Command {
	return &Command{Name: "claim", Usage: "Claim an eligible node: assign it and start work", Args: "<id>", Run: runTicketClaim}
}

// runTicketClaim is the guided "I'm taking this" verb (ADR 0041): it verifies a
// node is eligible (todo + dependencies satisfied), sets the assignee, moves it
// to in_progress, and prints the brief plus the next step — composing the
// existing primitives, never a new authority.
func runTicketClaim(ctx *Context, args []string) error {
	fs := ctx.NewFlagSet("gw ticket claim")
	var assignee string
	fs.StringVar(&assignee, "actor", ownerActor, "assignee for the node (default: human.owner)")
	pos, err := parseFlags(fs, args)
	if err != nil {
		return err
	}
	if len(pos) < 1 {
		return &Error{Code: "invalid_args", Message: "usage: gw ticket claim <id> [--actor <id>]"}
	}
	id := pos[0]

	// Reads (eligibility guard + brief) go direct; a running coordinator and the
	// CLI reader share the WAL store, so the direct read stays coherent (ADR 0031).
	p, db, err := ctx.openStore()
	if err != nil {
		return err
	}
	defer db.Close()

	// The mutation prefers the coordinator when one is running (ADR 0031).
	store, closeStore, err := ctx.openTicketStore()
	if err != nil {
		return err
	}
	defer closeStore()
	if err := claimNode(db, store, id, assignee); err != nil {
		return err
	}

	if ctx.JSON {
		return ctx.PrintJSON(map[string]string{
			"id": id, "status": string(ticket.StatusInProgress), "assignee": assignee})
	}
	fmt.Fprintf(ctx.Stdout, "Claimed %s -> in_progress (assignee: %s)\n\n", id, assignee)
	if brief, err := contextbrief.Build(db, p, id, false); err == nil {
		renderBrief(ctx, brief)
	}
	fmt.Fprintf(ctx.Stdout,
		"\nNext: make your change and stage it, then: gw ticket transition %s review && gw ticket land %s\n", id, id)
	return nil
}

// claimNode performs the eligibility-guarded claim: it verifies id is eligible
// (todo + dependencies satisfied), sets the assignee, and transitions it to
// in_progress. db is the read store for the eligibility guard; store is the
// mutation transport (direct or coordinator). Both are satisfied by *sqlite.DB.
func claimNode(db *sqlite.DB, store ticketStore, id, assignee string) error {
	tk, err := db.GetTicket(id)
	if err != nil {
		return ticketError(err, id)
	}
	if tk.Status != ticket.StatusTodo {
		return &Error{Code: "not_claimable",
			Message: fmt.Sprintf("%s is %s; only todo nodes can be claimed", id, tk.Status)}
	}
	if blockers, err := unmetDeps(db, id); err != nil {
		return err
	} else if len(blockers) > 0 {
		return &Error{Code: "blocked",
			Message: fmt.Sprintf("%s is blocked by: %s", id, strings.Join(blockers, ", "))}
	}
	// Set the assignee first (UpdateTicket leaves status untouched), then transition.
	tk.Assignee = assignee
	if err := store.UpdateTicket(tk, ownerActor); err != nil {
		return &Error{Code: "claim_failed", Message: err.Error()}
	}
	if err := store.TransitionTicket(id, ticket.StatusInProgress, ownerActor); err != nil {
		if errors.Is(err, sqlite.ErrIllegalTransition) {
			return &Error{Code: "illegal_transition", Message: err.Error()}
		}
		return &Error{Code: "claim_failed", Message: err.Error()}
	}
	return nil
}

// unmetDeps returns the "id (status)" descriptors of id's dependencies that are
// not yet done — the reason a todo node is not eligible (ADR 0024/0041).
func unmetDeps(db *sqlite.DB, id string) ([]string, error) {
	depIDs, err := db.DependencyIDs(id)
	if err != nil {
		return nil, &Error{Code: "store_error", Message: err.Error()}
	}
	var blockers []string
	for _, depID := range depIDs {
		dep, err := db.GetTicket(depID)
		if err != nil {
			return nil, &Error{Code: "store_error", Message: err.Error()}
		}
		if !ticket.DependencyMet(dep.Status) {
			blockers = append(blockers, fmt.Sprintf("%s (%s)", dep.ID, dep.Status))
		}
	}
	return blockers, nil
}

// --- rendering ---

func renderTicket(ctx *Context, t *ticket.Ticket, dependsOn []string) {
	w := ctx.Stdout
	fmt.Fprintf(w, "%s  %s\n", t.ID, t.Title)
	fmt.Fprintf(w, "  kind:       %s\n", t.Kind)
	fmt.Fprintf(w, "  node_type:  %s\n", orDash(string(t.NodeType), "(untriaged)"))
	if t.WorkType != "" {
		fmt.Fprintf(w, "  work_type:  %s\n", t.WorkType)
	}
	fmt.Fprintf(w, "  status:     %s\n", t.Status)
	fmt.Fprintf(w, "  parent:     %s\n", orDash(t.ParentID, "-"))
	fmt.Fprintf(w, "  assignee:   %s\n", orDash(t.Assignee, "-"))
	if t.RequestedActor != "" {
		fmt.Fprintf(w, "  requested_actor: %s\n", t.RequestedActor)
	}
	if len(dependsOn) > 0 {
		fmt.Fprintf(w, "  depends_on: %s\n", strings.Join(dependsOn, ", "))
	}
	if t.Priority != nil {
		fmt.Fprintf(w, "  priority:   %g\n", *t.Priority)
	}
	if len(t.Labels) > 0 {
		fmt.Fprintf(w, "  labels:     %s\n", strings.Join(t.Labels, ", "))
	}
	if t.Description != "" {
		fmt.Fprintf(w, "  description: %s\n", t.Description)
	}
	if len(t.Acceptance) > 0 {
		fmt.Fprintln(w, "  acceptance:")
		for _, a := range t.Acceptance {
			fmt.Fprintf(w, "    - %s\n", a)
		}
	}
	fmt.Fprintf(w, "  created:    %s\n", t.CreatedAt)
	fmt.Fprintf(w, "  updated:    %s\n", t.UpdatedAt)
}

func renderTicketList(ctx *Context, tickets []*ticket.Ticket, emptyMsg string) {
	if len(tickets) == 0 {
		fmt.Fprintln(ctx.Stdout, emptyMsg)
		return
	}
	fmt.Fprintf(ctx.Stdout, "%-8s  %-11s  %-9s  %s\n", "ID", "STATUS", "TYPE", "TITLE")
	for _, t := range tickets {
		fmt.Fprintf(ctx.Stdout, "%-8s  %-11s  %-9s  %s\n",
			t.ID, t.Status, orDash(string(t.NodeType), "-"), t.Title)
	}
}

func orDash(s, dash string) string {
	if s == "" {
		return dash
	}
	return s
}

// validatePriority enforces the [0,1] value range (ADR 0039).
func validatePriority(p float64) error {
	if p < 0 || p > 1 {
		return &Error{Code: "invalid_args", Message: fmt.Sprintf("priority %g out of range; must be in [0,1]", p)}
	}
	return nil
}

// ticketError maps a store error to a CLI error, distinguishing not-found.
func ticketError(err error, id string) error {
	if err == sqlite.ErrNotFound {
		return &Error{Code: "not_found", Message: fmt.Sprintf("ticket %q not found", id)}
	}
	return &Error{Code: "store_error", Message: err.Error()}
}
