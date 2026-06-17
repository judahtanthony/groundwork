// Package scaffold creates the .groundwork directory tree for `gw init`
// (see docs/contracts/file-layout.md). It writes durable, committed files only;
// the SQLite database is created lazily on first store use, not here.
package scaffold

import (
	"os"
	"path/filepath"

	"groundwork/internal/config"
)

// Result reports the outcome of an Init.
type Result struct {
	// AlreadyInitialized is true when a config.yaml was already present and Init
	// made no changes (idempotent, non-clobbering).
	AlreadyInitialized bool
	// Created lists paths written, relative to the project root.
	Created []string
}

// Init scaffolds .groundwork under root. If root already contains an
// initialized .groundwork (a config.yaml), Init makes no changes and reports
// AlreadyInitialized — re-running is safe and never clobbers existing state.
func Init(root string) (*Result, error) {
	p := &config.Project{Root: root}

	if _, err := os.Stat(p.ConfigPath()); err == nil {
		return &Result{AlreadyInitialized: true}, nil
	} else if !os.IsNotExist(err) {
		return nil, err
	}

	res := &Result{}

	// Directories. Each gets a .gitkeep so empty committed dirs survive.
	dirs := []string{
		p.Dir(),
		p.PoliciesDir(),
		p.SopsDir(),
		p.TicketsDir(),
	}
	for _, d := range dirs {
		if err := os.MkdirAll(d, 0o755); err != nil {
			return nil, err
		}
	}
	for _, d := range []string{p.SopsDir(), p.TicketsDir()} {
		if err := writeFile(root, filepath.Join(d, ".gitkeep"), nil, res); err != nil {
			return nil, err
		}
	}

	// config.yaml from defaults.
	cfgBytes, err := config.Marshal(ptr(config.Defaults()))
	if err != nil {
		return nil, err
	}
	if err := writeFile(root, p.ConfigPath(), cfgBytes, res); err != nil {
		return nil, err
	}

	// WORKFLOW.md and starter policies (committed canon).
	if err := writeFile(root, p.WorkflowPath(), []byte(workflowTemplate), res); err != nil {
		return nil, err
	}
	policyFiles := map[string]string{
		"trust.yaml":      trustPolicyTemplate,
		"validation.yaml": validationPolicyTemplate,
		"autonomy.yaml":   autonomyPolicyTemplate,
	}
	for name, body := range policyFiles {
		if err := writeFile(root, filepath.Join(p.PoliciesDir(), name), []byte(body), res); err != nil {
			return nil, err
		}
	}

	// Ignore guidance: a .gitignore inside .groundwork covering runtime tiers
	// (ADR 0007, ADR 0012). Paths are relative to .groundwork.
	if err := writeFile(root, filepath.Join(p.Dir(), ".gitignore"), []byte(gitignoreTemplate), res); err != nil {
		return nil, err
	}

	return res, nil
}

// writeFile writes data to path (which must be under root) and records it,
// relative to root, in res.Created.
func writeFile(root, path string, data []byte, res *Result) error {
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return err
	}
	rel, err := filepath.Rel(root, path)
	if err != nil {
		rel = path
	}
	res.Created = append(res.Created, rel)
	return nil
}

func ptr[T any](v T) *T { return &v }

const workflowTemplate = `# Workflow

This is the global operating policy for Groundwork in this repository. It is
committed canon and applies to every node. Task-type specifics live in
` + "`.groundwork/sops/<task-type>/`" + ` rather than here.

## Operating Notes

- Triage every claimed node as leaf or composite before working it.
- Keep leaf nodes to one verifiable change.
- Land only through validated, gated changes; human approval is required in v1.
`

const gitignoreTemplate = `# Groundwork runtime state — ignored by default (ADR 0007, ADR 0012).
state.sqlite
state.sqlite-wal
state.sqlite-shm
runs/
approvals/
views/
worktrees/
`

const trustPolicyTemplate = `schema: groundwork_trust_policy/v1
auto_approve:
  - id: internal_docs
    description: Allow documentation-only changes to internal agent guidance.
    when:
      files:
        - AGENTS.md
        - MEMORY.md
        - ".groundwork/**/*.md"
      change_type: documentation
      max_diff_lines: 200
require_human:
  - id: secrets
    files:
      - "**/.env*"
      - "**/*secret*"
  - id: landing_to_main_v1
    action_types: [land_to_main]
  - id: decomposition_v1
    action_types: [decompose]
`

const validationPolicyTemplate = `schema: groundwork_validation_policy/v1
templates:
  documentation:
    match:
      files: ["**/*.md", "AGENTS.md", "MEMORY.md"]
    required: []
    landing_risk_floor: low
  go:
    match:
      files: ["**/*.go"]
    required:
      - name: go_tests
        command: "go test ./..."
`

const autonomyPolicyTemplate = `schema: groundwork_autonomy_policy/v1
actions:
  execute:
    default: require_human
  land_to_main:
    default: require_human
  decompose:
    default: require_human
`
