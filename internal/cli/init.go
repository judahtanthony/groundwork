package cli

import (
	"fmt"
	"os"

	"groundwork/internal/scaffold"
)

func newInitCmd() *Command {
	return &Command{
		Name:  "init",
		Usage: "Initialize .groundwork in this repository",
		Run:   runInit,
	}
}

func runInit(ctx *Context, args []string) error {
	fs := ctx.NewFlagSet("gw init")
	if err := fs.Parse(args); err != nil {
		return err
	}

	root := ctx.RootFlag
	if root == "" {
		root = os.Getenv("GW_ROOT")
	}
	if root == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return &Error{Code: "cwd_failed", Message: fmt.Sprintf("determining working directory: %v", err)}
		}
		root = cwd
	}

	res, err := scaffold.Init(root)
	if err != nil {
		return &Error{Code: "init_failed", Message: err.Error()}
	}

	if ctx.JSON {
		return ctx.PrintJSON(map[string]any{
			"root":                root,
			"already_initialized": res.AlreadyInitialized,
			"created":             res.Created,
		})
	}

	if res.AlreadyInitialized {
		fmt.Fprintf(ctx.Stdout, "Groundwork is already initialized in %s\n", root)
		return nil
	}
	fmt.Fprintf(ctx.Stdout, "Initialized Groundwork in %s\n", root)
	for _, f := range res.Created {
		fmt.Fprintf(ctx.Stdout, "  created %s\n", f)
	}
	return nil
}
