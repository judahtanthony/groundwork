package cli

import (
	"flag"
	"strings"

	"groundwork/internal/config"
	"groundwork/internal/store/sqlite"
)

// actor is the audit actor for CLI-initiated mutations. CLI commands act on
// behalf of the local human owner; this is the default actor id from the
// scaffolded registry (ADR 0023, .groundwork/actors.yaml).
const ownerActor = "human.owner"

// openStore discovers the project, opens (lazily creating) the SQLite store,
// and runs migrations. Callers must Close the returned DB. It is the direct
// store path used by reads and by store-safe mutations when no coordinator is
// running (ADR 0031).
func (ctx *Context) openStore() (*config.Project, *sqlite.DB, error) {
	p, err := ctx.resolveProject()
	if err != nil {
		return nil, nil, err
	}
	db, err := openDB(p)
	if err != nil {
		return nil, nil, err
	}
	// Filesystem write-through (ADR 0053): direct CLI mutations rewrite the
	// affected ticket sidecars so files stay the source of truth even without a
	// running coordinator.
	db.SetExportDir(p.TicketsDir())
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
