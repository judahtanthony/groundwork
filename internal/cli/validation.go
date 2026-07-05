package cli

import (
	"fmt"
	"os"
	"os/exec"

	"groundwork/internal/git"
	"groundwork/internal/policy"
	"groundwork/internal/store/sqlite"
)

// newValidationCmd builds the `gw validation` subtree. Listing reads recorded
// results (store-safe); running executes the configured validation commands in
// the project root and records the outcomes.
func newValidationCmd() *Command {
	return &Command{
		Name:  "validation",
		Usage: "List and run validation checks",
		Sub: []*Command{
			{Name: "list", Usage: "List recorded validation results", Args: "<ticket-id>", Run: runValidationList},
			{Name: "run", Usage: "Run configured validation commands and record results", Args: "<ticket-id>", Run: runValidationRun},
		},
	}
}

func runValidationList(ctx *Context, args []string) error {
	fs := ctx.NewFlagSet("gw validation list")
	pos, err := parseFlags(fs, args)
	if err != nil {
		return err
	}
	if len(pos) < 1 {
		return &Error{Code: "invalid_args", Message: "usage: gw validation list <ticket-id>"}
	}
	_, db, err := ctx.openStore()
	if err != nil {
		return err
	}
	defer db.Close()

	results, err := db.ListValidationsForTicket(pos[0])
	if err != nil {
		return &Error{Code: "list_failed", Message: err.Error()}
	}
	if ctx.JSON {
		return ctx.PrintJSON(results)
	}
	if len(results) == 0 {
		fmt.Fprintln(ctx.Stdout, "No validation results.")
		return nil
	}
	fmt.Fprintf(ctx.Stdout, "%-8s  %-18s  %-8s  %s\n", "ID", "NAME", "STATUS", "COMMAND")
	for _, r := range results {
		fmt.Fprintf(ctx.Stdout, "%-8s  %-18s  %-8s  %s\n", r.ID, r.Name, r.Status, r.Command)
	}
	return nil
}

// runValidationRun executes the commands from validation templates in the
// project root and records each result through the coordinator (ADR 0031:
// recording is a coordinator-required mutation, so the running server's state
// and SSE stream stay coherent). In M2 (no per-node diff) it runs every template
// that defines commands; Phase 4 scopes this to the run's changed files.
func runValidationRun(ctx *Context, args []string) error {
	fs := ctx.NewFlagSet("gw validation run")
	pos, err := parseFlags(fs, args)
	if err != nil {
		return err
	}
	if len(pos) < 1 {
		return &Error{Code: "invalid_args", Message: "usage: gw validation run <ticket-id>"}
	}
	ticketID := pos[0]

	p, err := ctx.resolveProject()
	if err != nil {
		return err
	}
	c, err := ctx.requireCoordinator()
	if err != nil {
		return err
	}
	policies, _, err := policy.Load(p.PoliciesDir())
	if err != nil {
		return &Error{Code: "policy_error", Message: err.Error()}
	}
	if policies.Validation == nil {
		fmt.Fprintln(ctx.Stdout, "No validation policy configured.")
		return nil
	}

	ran := 0
	for name, tmpl := range policies.Validation.Templates {
		for _, check := range tmpl.Required {
			if check.Command == "" {
				continue
			}
			status := sqlite.ValidationPass
			cmd := exec.Command("sh", "-c", check.Command)
			cmd.Dir = p.Root
			if err := cmd.Run(); err != nil {
				status = sqlite.ValidationFail
			}
			if _, err := c.RecordValidation(ticketID, sqlite.ValidationResult{
				Name: check.Name, Command: check.Command, Status: status,
			}); err != nil {
				return &Error{Code: "record_failed", Message: err.Error()}
			}
			fmt.Fprintf(ctx.Stdout, "%s/%s: %s\n", name, check.Name, status)
			ran++
		}
	}
	if ran == 0 {
		fmt.Fprintln(ctx.Stdout, "No validation commands to run.")
	}
	return nil
}

// runTicketLand lands an approved node through the coordinator's validation gate
// (coordinator-required, ADR 0031).
func runTicketLand(ctx *Context, args []string) error {
	fs := ctx.NewFlagSet("gw ticket land")
	var override, all, preview, toParent bool
	fs.BoolVar(&override, "override", false, "land despite failing/missing validation (audited)")
	fs.BoolVar(&all, "all", false, "stage all changes before committing (like git commit -a)")
	fs.BoolVar(&preview, "preview", false, "show the staged diff that would be landed, without opening the approval")
	fs.BoolVar(&toParent, "to-parent", false, "land a child onto its root's integration branch (land_to_parent) instead of main")
	pos, err := parseFlags(fs, args)
	if err != nil {
		return err
	}
	if len(pos) < 1 {
		return &Error{Code: "invalid_args", Message: "usage: gw ticket land <id> [--all] [--override] [--preview] [--to-parent]"}
	}
	id := pos[0]

	// --preview is a read-only inspection of what the gate would commit: it shows
	// the staged diff (plus the ticket's regenerated export, added at commit time)
	// and opens no approval and contacts no coordinator (ADR 0041).
	if preview {
		p, perr := ctx.resolveProject()
		if perr != nil {
			return perr
		}
		repo, rerr := git.Open(p.Root)
		if rerr != nil {
			return &Error{Code: "not_a_repo", Message: "preview requires a git repository"}
		}
		return previewLanding(ctx, id, repo)
	}

	c, err := ctx.requireCoordinator()
	if err != nil {
		return err
	}

	// Decide the landing level. A child whose work lives on a run branch under a
	// root integration target must land to that branch (land_to_parent), not commit
	// the main working tree — otherwise the run's code is orphaned on gw/run/<id>
	// (ADR 0058). Auto-route unless the caller forced a level with --to-parent; a
	// route lookup failure falls through to the main-tree path, which surfaces any
	// real error itself.
	if !toParent {
		if route, _, rerr := c.LandRoute(id); rerr == nil && route == "parent" {
			toParent = true
		}
	}
	if toParent {
		branch, lerr := c.LandToParent(id)
		if lerr != nil {
			return &Error{Code: "land_failed", Message: lerr.Error()}
		}
		if ctx.JSON {
			return ctx.PrintJSON(map[string]any{"landed": true, "id": id, "landed_to": branch})
		}
		fmt.Fprintf(ctx.Stdout, "%s landed to %s\n", id, branch)
		return nil
	}

	// land_to_main: resolve what the landing commit will include before asking the
	// coordinator to land. The coordinator commits the git index (plus the
	// regenerated export); --all stages everything, and an empty index prompts to
	// include all (default yes). Staging here persists until the land/approve
	// completes the commit (ADR 0034). Skipped cleanly when the root is not a git
	// repo. (Not reached for land_to_parent, whose work is squashed from the run
	// branch, never staged from the main tree.)
	if p, perr := ctx.resolveProject(); perr == nil {
		if repo, rerr := git.Open(p.Root); rerr == nil {
			if err := resolveLandStaging(repo, all, os.Stdin, ctx.Stdout); err != nil {
				return &Error{Code: "stage_failed", Message: err.Error()}
			}
		}
	}

	res, err := c.LandTicket(id, override)
	if err != nil {
		return &Error{Code: "land_failed", Message: err.Error()}
	}
	if ctx.JSON {
		return ctx.PrintJSON(res)
	}
	if res.Landed {
		fmt.Fprintf(ctx.Stdout, "%s landed (%s)\n", res.Ticket.ID, res.Ticket.Status)
	} else {
		fmt.Fprintf(ctx.Stdout, "Landing requires approval %s; run \"gw approval approve %s\"\n",
			res.Approval.ID, res.Approval.ID)
	}
	return nil
}
