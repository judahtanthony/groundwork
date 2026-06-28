package cli

import (
	"errors"
	"fmt"

	"groundwork/internal/client"
	"groundwork/internal/store/sqlite"
)

// newApprovalCmd builds the `gw approval` subtree. Decisions are live coordinator
// actions (ADR 0031), so every subcommand requires a running coordinator.
func newApprovalCmd() *Command {
	return &Command{
		Name:  "approval",
		Usage: "List and decide approvals (requires the coordinator)",
		Sub: []*Command{
			{Name: "list", Usage: "List approvals", Run: runApprovalList},
			{Name: "show", Usage: "Show an approval", Args: "<approval-id>", Run: runApprovalShow},
			{Name: "approve", Usage: "Approve an approval", Args: "<approval-id>", Run: runApprovalApprove},
			{Name: "reject", Usage: "Reject an approval", Args: "<approval-id>", Run: runApprovalReject},
			{Name: "clarify", Usage: "Ask the agent to clarify", Args: "<approval-id>", Run: runApprovalClarify},
		},
	}
}

func runApprovalList(ctx *Context, args []string) error {
	fs := ctx.NewFlagSet("gw approval list")
	var status string
	fs.StringVar(&status, "status", "", "filter by status (pending, approved, rejected, clarifying)")
	if _, err := parseFlags(fs, args); err != nil {
		return err
	}
	c, err := ctx.requireCoordinator()
	if err != nil {
		return err
	}
	approvals, err := c.ListApprovals(status)
	if err != nil {
		return &Error{Code: "list_failed", Message: err.Error()}
	}
	if ctx.JSON {
		return ctx.PrintJSON(approvals)
	}
	if len(approvals) == 0 {
		fmt.Fprintln(ctx.Stdout, "No approvals.")
		return nil
	}
	fmt.Fprintf(ctx.Stdout, "%-8s  %-8s  %-13s  %-10s  %s\n", "ID", "TICKET", "TYPE", "STATUS", "SUMMARY")
	for _, a := range approvals {
		fmt.Fprintf(ctx.Stdout, "%-8s  %-8s  %-13s  %-10s  %s\n", a.ID, a.TicketID, a.Type, a.Status, a.Summary)
	}
	return nil
}

func runApprovalShow(ctx *Context, args []string) error {
	pos, err := positional(ctx, "gw approval show", args, 1, "usage: gw approval show <approval-id>")
	if err != nil {
		return err
	}
	c, err := ctx.requireCoordinator()
	if err != nil {
		return err
	}
	a, err := c.GetApproval(pos[0])
	if err != nil {
		return approvalError(err, pos[0])
	}
	if ctx.JSON {
		return ctx.PrintJSON(a)
	}
	w := ctx.Stdout
	fmt.Fprintf(w, "%s  ticket=%s  type=%s  status=%s\n", a.ID, a.TicketID, a.Type, a.Status)
	fmt.Fprintf(w, "  risk:       %s\n", a.RiskClass)
	fmt.Fprintf(w, "  summary:    %s\n", a.Summary)
	fmt.Fprintf(w, "  requested:  %s\n", a.RequestedByActor)
	if a.DecidedByActor != "" {
		fmt.Fprintf(w, "  decided_by: %s\n", a.DecidedByActor)
	}
	return nil
}

func runApprovalApprove(ctx *Context, args []string) error {
	return approvalDecision(ctx, "gw approval approve", args, "approve")
}
func runApprovalReject(ctx *Context, args []string) error {
	return approvalDecision(ctx, "gw approval reject", args, "reject")
}
func runApprovalClarify(ctx *Context, args []string) error {
	return approvalDecision(ctx, "gw approval clarify", args, "clarify")
}

func approvalDecision(ctx *Context, usage string, args []string, op string) error {
	fs := ctx.NewFlagSet(usage)
	var reason string
	fs.StringVar(&reason, "reason", "", "decision reason")
	pos, err := parseFlags(fs, args)
	if err != nil {
		return err
	}
	if len(pos) < 1 {
		return &Error{Code: "invalid_args", Message: "usage: " + usage + " <approval-id> [--reason ...]"}
	}
	c, err := ctx.requireCoordinator()
	if err != nil {
		return err
	}
	a, err := c.DecideApproval(pos[0], op, reason)
	if err != nil {
		return approvalError(err, pos[0])
	}
	if ctx.JSON {
		return ctx.PrintJSON(a)
	}
	fmt.Fprintf(ctx.Stdout, "%s -> %s\n", a.ID, a.Status)
	return nil
}

// runTicketDecompose and runTicketEscalate live with the ticket subtree but are
// coordinator-required (they open planning runs / route re-plan decisions).
func runTicketDecompose(ctx *Context, args []string) error {
	fs := ctx.NewFlagSet("gw ticket decompose")
	var children stringSlice
	var contract string
	fs.Var(&children, "child", "proposed child title (repeatable)")
	fs.StringVar(&contract, "contract", "", "parent contract as JSON")
	pos, err := parseFlags(fs, args)
	if err != nil {
		return err
	}
	if len(pos) < 1 || len(children) == 0 {
		return &Error{Code: "invalid_args", Message: "usage: gw ticket decompose <id> --child <title> [--child ...] [--contract <json>]"}
	}
	c, err := ctx.requireCoordinator()
	if err != nil {
		return err
	}
	specs := make([]sqlite.ChildSpec, 0, len(children))
	for _, title := range children {
		specs = append(specs, sqlite.ChildSpec{Title: title})
	}
	appr, childIDs, err := c.DecomposeTicket(pos[0], contract, specs)
	if err != nil {
		return approvalError(err, pos[0])
	}
	if ctx.JSON {
		return ctx.PrintJSON(map[string]any{"approval": appr, "child_ids": childIDs})
	}
	fmt.Fprintf(ctx.Stdout, "Proposed %d children for %s; approval %s pending\n", len(childIDs), pos[0], appr.ID)
	return nil
}

func runTicketEscalate(ctx *Context, args []string) error {
	fs := ctx.NewFlagSet("gw ticket escalate")
	var reason string
	fs.StringVar(&reason, "reason", "", "escalation reason")
	pos, err := parseFlags(fs, args)
	if err != nil {
		return err
	}
	if len(pos) < 1 {
		return &Error{Code: "invalid_args", Message: "usage: gw ticket escalate <id> [--reason ...]"}
	}
	c, err := ctx.requireCoordinator()
	if err != nil {
		return err
	}
	appr, err := c.EscalateTicket(pos[0], reason)
	if err != nil {
		return approvalError(err, pos[0])
	}
	if ctx.JSON {
		return ctx.PrintJSON(appr)
	}
	fmt.Fprintf(ctx.Stdout, "Escalated %s; re-plan approval %s pending\n", pos[0], appr.ID)
	return nil
}

// runTicketDecision raises a consequential decision work node for a blocked
// ticket (ADR 0052); the decision routes by work type and the blocked ticket
// depends on it.
func runTicketDecision(ctx *Context, args []string) error {
	fs := ctx.NewFlagSet("gw ticket decision")
	var workType, statement, actor, title string
	var acceptance stringSlice
	fs.StringVar(&workType, "work-type", "", "work type that routes the decision (e.g. architecture_decision)")
	fs.StringVar(&statement, "statement", "", "the question to decide")
	fs.StringVar(&title, "title", "", "decision node title (defaults from the statement)")
	fs.StringVar(&actor, "actor", "", "preferred actor for the decision")
	fs.Var(&acceptance, "acceptance", "decision acceptance criterion (repeatable)")
	pos, err := parseFlags(fs, args)
	if err != nil {
		return err
	}
	if len(pos) < 1 || workType == "" || statement == "" {
		return &Error{Code: "invalid_args", Message: "usage: gw ticket decision <id> --work-type <wt> --statement <q> [--title ...] [--actor ...] [--acceptance ...]"}
	}
	c, err := ctx.requireCoordinator()
	if err != nil {
		return err
	}
	res, err := c.RaiseDecision(pos[0], client.RaiseDecisionParams{
		Title: title, WorkType: workType, RequestedActor: actor, Statement: statement, Acceptance: acceptance,
	})
	if err != nil {
		return approvalError(err, pos[0])
	}
	if ctx.JSON {
		return ctx.PrintJSON(res)
	}
	fmt.Fprintf(ctx.Stdout, "Raised decision %s; %s now blocked on it\n", res.DecisionTicket, res.BlockedTicket)
	return nil
}

// runTicketInput records a bounded local input request without creating a work
// node (ADR 0052).
func runTicketInput(ctx *Context, args []string) error {
	fs := ctx.NewFlagSet("gw ticket input")
	var statement string
	fs.StringVar(&statement, "statement", "", "the clarification needed to continue")
	pos, err := parseFlags(fs, args)
	if err != nil {
		return err
	}
	if len(pos) < 1 || statement == "" {
		return &Error{Code: "invalid_args", Message: "usage: gw ticket input <id> --statement <q>"}
	}
	c, err := ctx.requireCoordinator()
	if err != nil {
		return err
	}
	rec, err := c.RequestInput(pos[0], statement, "")
	if err != nil {
		return approvalError(err, pos[0])
	}
	if ctx.JSON {
		return ctx.PrintJSON(rec)
	}
	fmt.Fprintf(ctx.Stdout, "Recorded input request on %s (seq %d)\n", pos[0], rec.Sequence)
	return nil
}

func approvalError(err error, id string) error {
	if errors.Is(err, sqlite.ErrNotFound) {
		return &Error{Code: "not_found", Message: fmt.Sprintf("approval %q not found", id)}
	}
	return &Error{Code: "approval_failed", Message: err.Error()}
}
