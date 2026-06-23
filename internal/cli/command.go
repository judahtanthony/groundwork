// Package cli implements the gw command tree on top of the standard-library
// flag package with a small hand-rolled subcommand router (see ADR 0016).
package cli

import (
	"fmt"
	"io"
	"strings"
)

// Command is a node in the gw command tree. A command either executes work
// (Run != nil), groups subcommands (len(Sub) > 0), or both. Pure groups (Run
// == nil) print their help when invoked without a matching subcommand.
type Command struct {
	// Name is the single token used to select this command from its parent.
	Name string
	// Usage is the one-line description shown in the parent's command list.
	Usage string
	// Args is an optional argument synopsis shown in help, e.g. "<id>".
	Args string
	// Run executes a leaf command. It receives the already-resolved context
	// and the arguments that follow this command's name.
	Run func(ctx *Context, args []string) error
	// Sub holds child commands, in display order.
	Sub []*Command
	// Flags documents a leaf command's flags for help output. It is descriptive
	// only — the values are still parsed in Run — mirroring how Usage and Args
	// are doc-only fields. The universal --json flag is rendered automatically.
	Flags []FlagDoc
}

// FlagDoc is one flag's help entry: the left-column spelling (e.g.
// "--status <status>") and its description.
type FlagDoc struct {
	Name string
	Desc string
}

// lookup returns the named subcommand, or nil.
func (c *Command) lookup(name string) *Command {
	for _, s := range c.Sub {
		if s.Name == name {
			return s
		}
	}
	return nil
}

// path is the space-joined command chain used in help and error messages.
func (c *Command) path(parents []string) string {
	return strings.Join(append(parents, c.Name), " ")
}

// dispatch walks the command tree. parents is the chain of command names that
// led here (including "gw"), used only for help and error text.
func (c *Command) dispatch(ctx *Context, parents []string, args []string) error {
	// An explicit help request at this level prints this command's help.
	if len(args) > 0 && isHelpArg(args[0]) {
		c.printHelp(ctx.Stdout, parents)
		return nil
	}

	if len(args) > 0 {
		if sub := c.lookup(args[0]); sub != nil {
			return sub.dispatch(ctx, append(parents, c.Name), args[1:])
		}
		// A non-flag token that matches no subcommand of a group is an error.
		if c.Run == nil && !strings.HasPrefix(args[0], "-") {
			return &Error{
				Code:    "unknown_command",
				Message: fmt.Sprintf("unknown command %q for %q", args[0], c.path(parents)),
			}
		}
	}

	if c.Run != nil {
		return c.Run(ctx, args)
	}

	// Group invoked with no subcommand: show its help.
	c.printHelp(ctx.Stdout, parents)
	return nil
}

// printHelp renders help for this command to w.
func (c *Command) printHelp(w io.Writer, parents []string) {
	full := c.path(parents)
	if c.Usage != "" {
		fmt.Fprintf(w, "%s — %s\n\n", full, c.Usage)
	}
	if len(c.Sub) > 0 {
		fmt.Fprintf(w, "Usage:\n  %s <command> [args]\n\nCommands:\n", full)
		width := 0
		for _, s := range c.Sub {
			if len(s.Name) > width {
				width = len(s.Name)
			}
		}
		for _, s := range c.Sub {
			fmt.Fprintf(w, "  %-*s  %s\n", width, s.Name, s.Usage)
		}
		fmt.Fprintf(w, "\nRun \"%s <command> -h\" for command help.\n", full)
		return
	}
	synopsis := full
	if c.Args != "" {
		synopsis += " " + c.Args
	}
	if len(c.Flags) == 0 {
		fmt.Fprintf(w, "Usage:\n  %s [--json]\n", synopsis)
		return
	}
	// A leaf with documented flags prints them in an aligned Flags section, plus
	// the universal --json flag every command accepts (ADR 0041).
	fmt.Fprintf(w, "Usage:\n  %s [flags]\n\nFlags:\n", synopsis)
	width := len("--json")
	for _, f := range c.Flags {
		if len(f.Name) > width {
			width = len(f.Name)
		}
	}
	for _, f := range c.Flags {
		fmt.Fprintf(w, "  %-*s  %s\n", width, f.Name, f.Desc)
	}
	fmt.Fprintf(w, "  %-*s  %s\n", width, "--json", "output machine-readable JSON")
}

// isHelpArg reports whether arg is a help request token.
func isHelpArg(arg string) bool {
	switch arg {
	case "help", "-h", "--help":
		return true
	}
	return false
}
