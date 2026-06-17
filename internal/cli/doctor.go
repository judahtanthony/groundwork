package cli

import (
	"fmt"
	"os"

	"groundwork/internal/config"
	"groundwork/internal/store/sqlite"
)

func newDoctorCmd() *Command {
	return &Command{Name: "doctor", Usage: "Diagnose the Groundwork environment", Run: runDoctor}
}

// checkStatus is one diagnostic outcome level.
type checkStatus string

const (
	checkOK    checkStatus = "ok"
	checkWarn  checkStatus = "warn"
	checkError checkStatus = "error"
)

type check struct {
	Name   string      `json:"name"`
	Status checkStatus `json:"status"`
	Detail string      `json:"detail"`
}

func runDoctor(ctx *Context, args []string) error {
	fs := ctx.NewFlagSet("gw doctor")
	if _, err := parseFlags(fs, args); err != nil {
		return err
	}

	var checks []check
	healthy := true
	add := func(name string, st checkStatus, detail string) {
		checks = append(checks, check{name, st, detail})
		if st == checkError {
			healthy = false
		}
	}

	cwd, err := os.Getwd()
	if err != nil {
		return &Error{Code: "cwd_failed", Message: err.Error()}
	}

	p, err := config.Open(cwd, ctx.RootFlag)
	if err != nil {
		add("project", checkError, err.Error())
		return reportDoctor(ctx, checks, healthy)
	}
	add("project", checkOK, "root: "+p.Root)
	add("config", checkOK, fmt.Sprintf("schema: %s, runtime: %s", p.Config.Schema, p.Config.Runtime))
	for _, w := range p.Warnings {
		add("config", checkWarn, w)
	}

	// Database: report status without forcing creation.
	if _, statErr := os.Stat(p.DBPath()); os.IsNotExist(statErr) {
		add("database", checkOK, "not created yet (created lazily on first store use)")
	} else {
		db, derr := sqlite.Open(p.DBPath())
		if derr != nil {
			add("database", checkError, derr.Error())
		} else {
			defer db.Close()
			if merr := db.Migrate(); merr != nil {
				add("database", checkError, merr.Error())
			} else {
				ids, _ := db.AppliedMigrationIDs()
				latest := "none"
				if len(ids) > 0 {
					latest = ids[len(ids)-1]
				}
				tickets, _ := db.ListTickets()
				add("database", checkOK, fmt.Sprintf("%s, migrations applied: %d (latest %s), tickets: %d",
					p.DBPath(), len(ids), latest, len(tickets)))
			}
		}
	}

	return reportDoctor(ctx, checks, healthy)
}

func reportDoctor(ctx *Context, checks []check, healthy bool) error {
	if ctx.JSON {
		if err := ctx.PrintJSON(map[string]any{"healthy": healthy, "checks": checks}); err != nil {
			return err
		}
	} else {
		for _, c := range checks {
			fmt.Fprintf(ctx.Stdout, "[%-5s] %-9s %s\n", c.Status, c.Name, c.Detail)
		}
	}
	if !healthy {
		// Signal failure via exit code without printing a duplicate error line.
		return &silentError{}
	}
	return nil
}

// silentError carries a non-zero exit code without producing error output
// (doctor already printed its findings).
type silentError struct{}

func (*silentError) Error() string { return "" }
