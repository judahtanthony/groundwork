package cli

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
)

// Context carries the resolved global state and IO streams through the command
// tree. It is created once per invocation in Main.
type Context struct {
	Stdout io.Writer
	Stderr io.Writer

	// JSON selects machine-readable output. It may be set globally
	// (gw --json ...) or per-command (gw ticket list --json).
	JSON bool

	// RootFlag holds the value of the global --root flag (project root
	// override). Empty means "discover from the working directory" (ADR 0021).
	RootFlag string
}

// NewFlagSet returns a FlagSet pre-wired with the common --json flag, bound to
// ctx.JSON and defaulting to its current (possibly globally set) value. Leaf
// commands use this so --json works both before and after the subcommand.
func (ctx *Context) NewFlagSet(name string) *flag.FlagSet {
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	fs.SetOutput(ctx.Stderr)
	fs.BoolVar(&ctx.JSON, "json", ctx.JSON, "output machine-readable JSON")
	return fs
}

// PrintJSON writes v as indented JSON followed by a newline to stdout.
func (ctx *Context) PrintJSON(v any) error {
	enc := json.NewEncoder(ctx.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(v); err != nil {
		return &Error{Code: "encode_failed", Message: fmt.Sprintf("encoding JSON output: %v", err)}
	}
	return nil
}
