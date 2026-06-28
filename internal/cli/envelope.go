package cli

import (
	"errors"
	"fmt"

	"groundwork/internal/envelope"
	"groundwork/internal/store/sqlite"
)

// newEnvelopeCmd is the `gw envelope` group (ADR 0054): inspect approval
// envelopes and revoke them. Approving an envelope goes through the normal
// approval flow (gw approval approve), so it stays human-gated.
func newEnvelopeCmd() *Command {
	return &Command{
		Name:  "envelope",
		Usage: "Propose, inspect, and revoke approval envelopes",
		Sub: []*Command{
			{Name: "propose", Usage: "Propose an envelope for a node (requires the coordinator)", Args: "<node-id>", Run: runEnvelopePropose, Flags: []FlagDoc{
				{"--action", "approved action: decompose_children|execute_children|land_children_to_parent|replan_within_goal (repeatable)"},
				{"--role", "allowed actor role (repeatable)"},
				{"--work-type", "allowed work type for planning (repeatable)"},
				{"--allow", "scope: allowed file glob (repeatable)"},
				{"--deny", "scope: denied file glob (repeatable)"},
				{"--risk-ceiling", "max risk class the envelope authorizes (low|medium|high)"},
				{"--max-depth", "max decomposition depth"},
				{"--max-children", "max children per node"},
			}},
			{Name: "list", Usage: "List envelopes", Run: runEnvelopeList, Flags: []FlagDoc{
				{"--status", "filter by status (active|revoked|superseded)"},
			}},
			{Name: "show", Usage: "Show the active envelope for a node", Args: "<node-id>", Run: runEnvelopeShow},
			{Name: "revoke", Usage: "Revoke an envelope by id", Args: "<envelope-id>", Run: runEnvelopeRevoke},
		},
	}
}

// runEnvelopePropose opens a human-gated approve_envelope approval carrying the
// drafted boundary, through the coordinator so the proposal flows through policy
// and the running server's state/SSE stay coherent (ADR 0054/0031). Approving the
// returned approval (gw approval approve) activates the envelope.
func runEnvelopePropose(ctx *Context, args []string) error {
	fs := ctx.NewFlagSet("gw envelope propose")
	var actions, roles, workTypes, allow, deny stringSlice
	var riskCeiling string
	var maxDepth, maxChildren int
	fs.Var(&actions, "action", "approved action (repeatable)")
	fs.Var(&roles, "role", "allowed actor role (repeatable)")
	fs.Var(&workTypes, "work-type", "allowed work type (repeatable)")
	fs.Var(&allow, "allow", "scope allow glob (repeatable)")
	fs.Var(&deny, "deny", "scope deny glob (repeatable)")
	fs.StringVar(&riskCeiling, "risk-ceiling", "", "max risk class (low|medium|high)")
	fs.IntVar(&maxDepth, "max-depth", 0, "max decomposition depth")
	fs.IntVar(&maxChildren, "max-children", 0, "max children per node")
	pos, err := parseFlags(fs, args)
	if err != nil {
		return err
	}
	if len(pos) != 1 || len(actions) == 0 {
		return &Error{Code: "invalid_args", Message: "usage: gw envelope propose <node-id> --action <action> [--action ...] [--role ...]"}
	}
	draft := &envelope.Envelope{
		ApprovedActions: actions,
		AllowedRoles:    roles,
		RiskCeiling:     riskCeiling,
		Planning:        envelope.Planning{MaxDepth: maxDepth, MaxChildren: maxChildren, AllowedWorkTypes: workTypes},
		Scope:           envelope.Scope{Files: envelope.FileScope{Allow: allow, Deny: deny}},
	}
	c, err := ctx.requireCoordinator()
	if err != nil {
		return err
	}
	appr, err := c.ProposeEnvelope(pos[0], draft)
	if err != nil {
		return approvalError(err, pos[0])
	}
	if ctx.JSON {
		return ctx.PrintJSON(appr)
	}
	fmt.Fprintf(ctx.Stdout, "Proposed envelope for %s — approval %s is pending.\n", pos[0], appr.ID)
	fmt.Fprintf(ctx.Stdout, "Approve it with: gw approval approve %s\n", appr.ID)
	return nil
}

func runEnvelopeList(ctx *Context, args []string) error {
	fs := ctx.NewFlagSet("gw envelope list")
	var status string
	fs.StringVar(&status, "status", "", "filter by status")
	if err := fs.Parse(args); err != nil {
		return err
	}
	_, db, err := ctx.openStore()
	if err != nil {
		return err
	}
	defer db.Close()
	envs, err := db.ListEnvelopes(status)
	if err != nil {
		return &Error{Code: "list_failed", Message: err.Error()}
	}
	if ctx.JSON {
		return ctx.PrintJSON(envs)
	}
	if len(envs) == 0 {
		fmt.Fprintln(ctx.Stdout, "No envelopes.")
		return nil
	}
	fmt.Fprintf(ctx.Stdout, "%-10s  %-10s  %-10s  %s\n", "ID", "NODE", "STATUS", "ACTIONS")
	for _, e := range envs {
		fmt.Fprintf(ctx.Stdout, "%-10s  %-10s  %-10s  %v\n", e.ID, e.NodeID, e.Status, e.ApprovedActions)
	}
	return nil
}

func runEnvelopeShow(ctx *Context, args []string) error {
	if len(args) != 1 {
		return &Error{Code: "invalid_args", Message: "usage: gw envelope show <node-id>"}
	}
	_, db, err := ctx.openStore()
	if err != nil {
		return err
	}
	defer db.Close()
	e, err := db.GetActiveEnvelopeForNode(args[0])
	if err != nil {
		return &Error{Code: "store_error", Message: err.Error()}
	}
	if e == nil {
		if ctx.JSON {
			return ctx.PrintJSON(nil)
		}
		fmt.Fprintf(ctx.Stdout, "No active envelope for %s.\n", args[0])
		return nil
	}
	if ctx.JSON {
		return ctx.PrintJSON(e)
	}
	fmt.Fprintf(ctx.Stdout, "Envelope %s for %s [%s]\n", e.ID, e.NodeID, e.Status)
	fmt.Fprintf(ctx.Stdout, "  approved by:   %s\n", e.ApprovedBy)
	fmt.Fprintf(ctx.Stdout, "  actions:       %v\n", e.ApprovedActions)
	fmt.Fprintf(ctx.Stdout, "  roles:         %v\n", e.AllowedRoles)
	fmt.Fprintf(ctx.Stdout, "  work types:    %v\n", e.Planning.AllowedWorkTypes)
	fmt.Fprintf(ctx.Stdout, "  risk ceiling:  %s\n", e.RiskCeiling)
	fmt.Fprintf(ctx.Stdout, "  scope allow:   %v\n", e.Scope.Files.Allow)
	fmt.Fprintf(ctx.Stdout, "  scope deny:    %v\n", e.Scope.Files.Deny)
	return nil
}

func runEnvelopeRevoke(ctx *Context, args []string) error {
	if len(args) != 1 {
		return &Error{Code: "invalid_args", Message: "usage: gw envelope revoke <envelope-id>"}
	}
	proj, db, err := ctx.openStore()
	if err != nil {
		return err
	}
	defer db.Close()
	e, err := db.GetEnvelope(args[0])
	if err != nil {
		if errors.Is(err, sqlite.ErrNotFound) {
			return &Error{Code: "not_found", Message: "envelope not found"}
		}
		return &Error{Code: "store_error", Message: err.Error()}
	}
	if err := db.SetEnvelopeStatus(e.ID, envelope.StatusRevoked); err != nil {
		return &Error{Code: "store_error", Message: err.Error()}
	}
	e.Status = envelope.StatusRevoked
	if err := envelope.Write(proj.TicketsDir(), e); err != nil {
		return &Error{Code: "sidecar_error", Message: err.Error()}
	}
	if ctx.JSON {
		return ctx.PrintJSON(e)
	}
	fmt.Fprintf(ctx.Stdout, "Revoked %s.\n", e.ID)
	return nil
}
