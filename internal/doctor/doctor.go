// Package doctor implements the project health checks shared by the CLI and
// coordinator settings API.
package doctor

import (
	"fmt"
	"os"

	"groundwork/internal/actor"
	"groundwork/internal/config"
	"groundwork/internal/store/sqlite"
)

type Status string

const (
	OK    Status = "ok"
	Warn  Status = "warn"
	Error Status = "error"
)

type Check struct {
	Name   string `json:"name"`
	Status Status `json:"status"`
	Detail string `json:"detail"`
}

type Report struct {
	Healthy bool    `json:"healthy"`
	Checks  []Check `json:"checks"`
}

// Run evaluates the same project, config, actor-registry, and database checks
// surfaced by gw doctor.
func Run(p *config.Project) Report {
	report := Report{Healthy: true, Checks: []Check{}}
	add := func(name string, status Status, detail string) {
		report.Checks = append(report.Checks, Check{Name: name, Status: status, Detail: detail})
		if status == Error {
			report.Healthy = false
		}
	}

	add("project", OK, "root: "+p.Root)
	add("config", OK, fmt.Sprintf("schema: %s, runtime: %s", p.Config.Schema, p.Config.Runtime))
	for _, warning := range p.Warnings {
		add("config", Warn, warning)
	}

	if _, err := os.Stat(p.ActorsPath()); os.IsNotExist(err) {
		add("actors", Warn, `no actors.yaml (run "gw init" to scaffold one)`)
	} else if registry, warnings, err := actor.Load(p.ActorsPath()); err != nil {
		add("actors", Error, err.Error())
	} else {
		for _, warning := range warnings {
			add("actors", Warn, warning)
		}
		add("actors", OK, fmt.Sprintf("%d actors", len(registry.Actors)))
	}

	if _, err := os.Stat(p.DBPath()); os.IsNotExist(err) {
		add("database", OK, "not created yet (created lazily on first store use)")
	} else {
		db, err := sqlite.Open(p.DBPath())
		if err != nil {
			add("database", Error, err.Error())
		} else {
			defer db.Close()
			if err := db.Migrate(); err != nil {
				add("database", Error, err.Error())
			} else {
				ids, _ := db.AppliedMigrationIDs()
				latest := "none"
				if len(ids) > 0 {
					latest = ids[len(ids)-1]
				}
				tickets, _ := db.ListTickets()
				add("database", OK, fmt.Sprintf("%s, migrations applied: %d (latest %s), tickets: %d",
					p.DBPath(), len(ids), latest, len(tickets)))
			}
		}
	}
	return report
}
