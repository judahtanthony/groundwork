package cli

import (
	"errors"
	"flag"
	"os"
	"strings"

	"groundwork/internal/config"
	"groundwork/internal/store/sqlite"
)

// actor is the audit actor for CLI-initiated mutations in Phase 1. All gw CLI
// commands act on behalf of a human operator.
const actor = "human"

// openStore discovers the project, opens (lazily creating) the SQLite store,
// and runs migrations. Callers must Close the returned DB.
func (ctx *Context) openStore() (*config.Project, *sqlite.DB, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, nil, &Error{Code: "cwd_failed", Message: err.Error()}
	}
	p, err := config.Open(cwd, ctx.RootFlag)
	if err != nil {
		if errors.Is(err, config.ErrProjectNotFound) {
			return nil, nil, &Error{Code: "no_project", Message: err.Error()}
		}
		return nil, nil, &Error{Code: "config_error", Message: err.Error()}
	}
	for _, w := range p.Warnings {
		// Warnings are advisory (ADR 0021); surface them on stderr.
		if !ctx.JSON {
			ctx.Stderr.Write([]byte("gw: warning: " + w + "\n"))
		}
	}
	db, err := sqlite.Open(p.DBPath())
	if err != nil {
		return nil, nil, &Error{Code: "store_error", Message: err.Error()}
	}
	if err := db.Migrate(); err != nil {
		db.Close()
		return nil, nil, &Error{Code: "migrate_error", Message: err.Error()}
	}
	return p, db, nil
}

// stringSlice is a repeatable string flag (e.g. --label a --label b).
type stringSlice []string

func (s *stringSlice) String() string {
	if s == nil {
		return ""
	}
	return strings.Join(*s, ", ")
}

func (s *stringSlice) Set(v string) error {
	*s = append(*s, v)
	return nil
}

// setFlags returns the set of flags explicitly provided on the command line.
func setFlags(fs *flag.FlagSet) map[string]bool {
	m := map[string]bool{}
	fs.Visit(func(f *flag.Flag) { m[f.Name] = true })
	return m
}

// parseFlags parses args allowing flags and positionals to be interspersed
// (the stdlib flag package otherwise stops at the first positional). It returns
// the collected positional arguments in order.
func parseFlags(fs *flag.FlagSet, args []string) ([]string, error) {
	var positionals []string
	for {
		if err := fs.Parse(args); err != nil {
			return nil, err
		}
		rest := fs.Args()
		if len(rest) == 0 {
			return positionals, nil
		}
		positionals = append(positionals, rest[0])
		args = rest[1:]
	}
}
