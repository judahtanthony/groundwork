package cli

import (
	"fmt"
	"os"

	"groundwork/internal/config"
	"groundwork/internal/doctor"
)

func newDoctorCmd() *Command {
	return &Command{Name: "doctor", Usage: "Diagnose the Groundwork environment", Run: runDoctor}
}

func runDoctor(ctx *Context, args []string) error {
	fs := ctx.NewFlagSet("gw doctor")
	if _, err := parseFlags(fs, args); err != nil {
		return err
	}

	cwd, err := os.Getwd()
	if err != nil {
		return &Error{Code: "cwd_failed", Message: err.Error()}
	}

	p, err := config.Open(cwd, ctx.RootFlag)
	if err != nil {
		return reportDoctor(ctx, doctor.Report{
			Healthy: false,
			Checks:  []doctor.Check{{Name: "project", Status: doctor.Error, Detail: err.Error()}},
		})
	}
	return reportDoctor(ctx, doctor.Run(p))
}

func reportDoctor(ctx *Context, report doctor.Report) error {
	if ctx.JSON {
		if err := ctx.PrintJSON(report); err != nil {
			return err
		}
	} else {
		for _, c := range report.Checks {
			fmt.Fprintf(ctx.Stdout, "[%-5s] %-9s %s\n", c.Status, c.Name, c.Detail)
		}
	}
	if !report.Healthy {
		// Signal failure via exit code without printing a duplicate error line.
		return &silentError{}
	}
	return nil
}

// silentError carries a non-zero exit code without producing error output
// (doctor already printed its findings).
type silentError struct{}

func (*silentError) Error() string { return "" }
