package cli

import (
	"fmt"
	"strings"

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
		priority                                    int
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
	fs.IntVar(&priority, "priority", 0, "priority")
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

	_, db, err := ctx.openStore()
	if err != nil {
		return err
	}
	defer db.Close()

	if err := db.CreateTicket(t, ownerActor); err != nil {
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
	fs.StringVar(&status, "status", "", "filter by status")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if status != "" && !ticket.Status(status).Valid() {
		return &Error{Code: "invalid_args", Message: fmt.Sprintf("invalid status %q", status)}
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
	renderTicketList(ctx, tickets)
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
		priority                    int
		labels, acceptance          stringSlice
	)
	fs.StringVar(&title, "title", "", "new title")
	fs.StringVar(&kind, "kind", "", "new advisory kind")
	fs.StringVar(&desc, "description", "", "new description")
	fs.StringVar(&assignee, "assignee", "", "new assignee (display-only label)")
	fs.StringVar(&workType, "work-type", "", "new work type")
	fs.StringVar(&requestedActor, "requested-actor", "", "new requested actor (routing hint)")
	fs.IntVar(&priority, "priority", 0, "new priority")
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

	_, db, err := ctx.openStore()
	if err != nil {
		return err
	}
	defer db.Close()

	t, err := db.GetTicket(pos[0])
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
		t.Priority = &priority
	}
	if set["label"] {
		t.Labels = labels
	}
	if set["acceptance"] {
		t.Acceptance = acceptance
	}

	if err := db.UpdateTicket(t, ownerActor); err != nil {
		return &Error{Code: "edit_failed", Message: err.Error()}
	}
	if ctx.JSON {
		return ctx.PrintJSON(t)
	}
	fmt.Fprintf(ctx.Stdout, "Updated %s\n", t.ID)
	return nil
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
		fmt.Fprintf(w, "  priority:   %d\n", *t.Priority)
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

func renderTicketList(ctx *Context, tickets []*ticket.Ticket) {
	if len(tickets) == 0 {
		fmt.Fprintln(ctx.Stdout, "No tickets.")
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

// ticketError maps a store error to a CLI error, distinguishing not-found.
func ticketError(err error, id string) error {
	if err == sqlite.ErrNotFound {
		return &Error{Code: "not_found", Message: fmt.Sprintf("ticket %q not found", id)}
	}
	return &Error{Code: "store_error", Message: err.Error()}
}
