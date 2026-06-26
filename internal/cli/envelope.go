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
		Usage: "Inspect and revoke approval envelopes",
		Sub: []*Command{
			{Name: "list", Usage: "List envelopes", Run: runEnvelopeList, Flags: []FlagDoc{
				{"--status", "filter by status (active|revoked|superseded)"},
			}},
			{Name: "show", Usage: "Show the active envelope for a node", Args: "<node-id>", Run: runEnvelopeShow},
			{Name: "revoke", Usage: "Revoke an envelope by id", Args: "<envelope-id>", Run: runEnvelopeRevoke},
		},
	}
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
