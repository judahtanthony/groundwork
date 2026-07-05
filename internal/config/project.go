package config

import (
	"errors"
	"os"
	"path/filepath"
)

// GroundworkDir is the managed dot-directory name (ADR 0003).
const GroundworkDir = ".groundwork"

// ErrProjectNotFound is returned when no .groundwork directory can be located.
var ErrProjectNotFound = errors.New("no .groundwork directory found; run \"gw init\" to create one")

// Project bundles a discovered project root with its resolved paths and (when
// loaded) its config. Root is the directory that contains .groundwork.
type Project struct {
	Root     string
	Config   *Config
	Warnings []string
}

// Discover locates the project root. If override is non-empty it is used
// directly (and must contain a .groundwork directory); otherwise Discover walks
// parent directories from startDir until it finds one. The GW_ROOT environment
// variable is consulted by Open, not here.
func Discover(startDir, override string) (string, error) {
	if override != "" {
		abs, err := filepath.Abs(override)
		if err != nil {
			return "", err
		}
		if isDir(filepath.Join(abs, GroundworkDir)) {
			return abs, nil
		}
		return "", ErrProjectNotFound
	}

	dir, err := filepath.Abs(startDir)
	if err != nil {
		return "", err
	}
	for {
		if isDir(filepath.Join(dir, GroundworkDir)) {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", ErrProjectNotFound
		}
		dir = parent
	}
}

// Open discovers the project root (honoring override, then GW_ROOT, then an
// upward walk from startDir) and loads its config.
func Open(startDir, override string) (*Project, error) {
	if override == "" {
		override = os.Getenv("GW_ROOT")
	}
	root, err := Discover(startDir, override)
	if err != nil {
		return nil, err
	}
	p := &Project{Root: root}

	data, err := os.ReadFile(p.ConfigPath())
	if err != nil {
		return nil, err
	}
	cfg, warnings, err := Parse(data)
	if err != nil {
		return nil, err
	}
	p.Config = cfg
	p.Warnings = warnings
	return p, nil
}

// Path helpers centralize the file-layout contract (docs/contracts/file-layout.md).

// Dir returns the absolute path to the project's .groundwork directory.
func (p *Project) Dir() string { return filepath.Join(p.Root, GroundworkDir) }

// ConfigPath returns the path to config.yaml.
func (p *Project) ConfigPath() string { return filepath.Join(p.Dir(), "config.yaml") }

// ActorsPath returns the path to the actor registry (ADR 0023).
func (p *Project) ActorsPath() string { return filepath.Join(p.Dir(), "actors.yaml") }

// DBPath returns the path to the operational SQLite database.
func (p *Project) DBPath() string { return filepath.Join(p.Dir(), "state.sqlite") }

// TicketsDir returns the ticket-export directory.
func (p *Project) TicketsDir() string { return filepath.Join(p.Dir(), "tickets") }

// RunsDir returns the run-log directory (ignored runtime state).
func (p *Project) RunsDir() string { return filepath.Join(p.Dir(), "runs") }

// WorktreesDir returns the per-run isolated worktree root (ignored runtime state,
// ADR 0059). Each run gets <WorktreesDir>/<run-id>.
func (p *Project) WorktreesDir() string { return filepath.Join(p.Dir(), "worktrees") }

// JournalDir returns the per-node decision-journal directory (tier-1 ephemeral,
// ignored; ADR 0013).
func (p *Project) JournalDir() string { return filepath.Join(p.RunsDir(), "journal") }

// SopsDir returns the SOP directory.
func (p *Project) SopsDir() string { return filepath.Join(p.Dir(), "sops") }

// PoliciesDir returns the policy directory.
func (p *Project) PoliciesDir() string { return filepath.Join(p.Dir(), "policies") }

// WorkflowPath returns the path to WORKFLOW.md.
func (p *Project) WorkflowPath() string { return filepath.Join(p.Dir(), "WORKFLOW.md") }

func isDir(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}
