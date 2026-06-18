package cli

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
)

// Error is a CLI error carrying a stable machine code. It renders as the JSON
// error envelope from docs/contracts/http-api.md when --json is set.
type Error struct {
	Code    string
	Message string
}

func (e *Error) Error() string { return e.Message }

// Main is the process entry point. It parses global flags, dispatches to the
// command tree, and returns the process exit code.
func Main(args []string) int {
	ctx := &Context{Stdout: os.Stdout, Stderr: os.Stderr}

	global := flag.NewFlagSet("gw", flag.ContinueOnError)
	global.SetOutput(ctx.Stderr)
	global.BoolVar(&ctx.JSON, "json", false, "output machine-readable JSON")
	global.StringVar(&ctx.RootFlag, "root", "", "project root override (default: discover .groundwork upward)")

	// flag stops at the first non-flag token, so global flags must precede the
	// subcommand: `gw --json ticket list`.
	if err := global.Parse(args[1:]); err != nil {
		// flag already printed the error; treat as usage failure.
		return 2
	}

	root := buildRoot()
	if err := root.dispatch(ctx, nil, global.Args()); err != nil {
		var silent *silentError
		if !errors.As(err, &silent) {
			printError(ctx, err)
		}
		return 1
	}
	return 0
}

// printError renders err to stderr, as a JSON envelope when --json is set.
func printError(ctx *Context, err error) {
	code := "error"
	var ce *Error
	if errors.As(err, &ce) {
		code = ce.Code
	}
	if ctx.JSON {
		writeErrorEnvelope(ctx.Stderr, code, err.Error())
		return
	}
	fmt.Fprintf(ctx.Stderr, "gw: %s\n", err.Error())
}

// writeErrorEnvelope writes {"error":{"code","message"}} to w.
func writeErrorEnvelope(w io.Writer, code, message string) {
	type envelope struct {
		Error struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	var e envelope
	e.Error.Code = code
	e.Error.Message = message
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	_ = enc.Encode(e)
}

// buildRoot assembles the gw command tree. Commands are filled in as their
// Phase 1 tickets land; unimplemented ones return a clear not_implemented error
// while keeping `gw help` complete.
func buildRoot() *Command {
	return &Command{
		Name:  "gw",
		Usage: "local-first coordination for coding agents",
		Sub: []*Command{
			newInitCmd(),
			{Name: "ticket", Usage: "Manage work-tree nodes (tickets)", Sub: ticketSubcommands()},
			newContextCmd(),
			newStatusCmd(),
			newBoardCmd(),
			newExportCmd(),
			newDoctorCmd(),
			newVersionCmd(),
		},
	}
}

// ticketSubcommands returns the `gw ticket` subtree. Each is wired to the store
// as its ticket lands.
func ticketSubcommands() []*Command {
	return []*Command{
		newTicketCreateCmd(),
		newTicketListCmd(),
		newTicketShowCmd(),
		newTicketEditCmd(),
		newTicketTransitionCmd(),
		newTicketTriageCmd(),
		newTicketTreeCmd(),
		newContextCmd(),
		newTicketLinkCmd(),
		newExportCmd(),
	}
}
