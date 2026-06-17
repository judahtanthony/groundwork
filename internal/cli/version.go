package cli

import "fmt"

// Version is the gw build version. It is overridable at build time via
// -ldflags "-X groundwork/internal/cli.Version=...".
var Version = "0.0.0-dev"

func newVersionCmd() *Command {
	return &Command{
		Name:  "version",
		Usage: "Print the gw version",
		Run: func(ctx *Context, args []string) error {
			if ctx.JSON {
				return ctx.PrintJSON(map[string]string{"version": Version})
			}
			fmt.Fprintln(ctx.Stdout, Version)
			return nil
		},
	}
}
