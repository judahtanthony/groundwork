// Package actor parses and validates the local actor registry
// (.groundwork/actors.yaml, ADR 0023). The registry is file-authoritative canon
// (ADR 0012/0013): it is never copied into SQLite as rows; runs snapshot the
// selected actor separately. Phase 1 provides parsing, validation, and the
// `gw actor` read commands; actor-aware routing and selection are Phase 2.
package actor

import (
	"fmt"
	"os"
	"sort"

	"gopkg.in/yaml.v3"
)

// SchemaVersion is the actor registry schema this build understands.
const SchemaVersion = "groundwork_actors/v1"

// Type is an actor category. These are policy inputs, not hard security
// boundaries by themselves (docs/architecture/actors.md).
type Type string

const (
	TypeHuman   Type = "human"
	TypeAIAgent Type = "ai_agent"
	TypeAIJudge Type = "ai_judge"
	TypeTool    Type = "tool"
)

func (t Type) valid() bool {
	switch t {
	case TypeHuman, TypeAIAgent, TypeAIJudge, TypeTool:
		return true
	}
	return false
}

// Capabilities are coarse routing/authorization claims (forward-compatible:
// later versions add tools, MCPs, skills, domains, etc.).
type Capabilities struct {
	WorkTypes []string `yaml:"work_types" json:"work_types,omitempty"`
	Approve   []string `yaml:"approve" json:"approve,omitempty"`
	Review    []string `yaml:"review" json:"review,omitempty"`
}

// Limits are optional risk/scope limits.
type Limits struct {
	MaxRiskClass string `yaml:"max_risk_class" json:"max_risk_class,omitempty"`
}

// Actor is one registry entry.
type Actor struct {
	ID           string       `yaml:"id" json:"id"`
	Type         Type         `yaml:"type" json:"type"`
	DisplayName  string       `yaml:"display_name" json:"display_name,omitempty"`
	Roles        []string     `yaml:"roles" json:"roles,omitempty"`
	Runtime      string       `yaml:"runtime" json:"runtime,omitempty"`
	Model        string       `yaml:"model" json:"model,omitempty"`
	Sandbox      string       `yaml:"sandbox" json:"sandbox,omitempty"`
	Capabilities Capabilities `yaml:"capabilities" json:"capabilities,omitempty"`
	Limits       Limits       `yaml:"limits" json:"limits,omitempty"`
}

// Registry is the parsed actors.yaml.
type Registry struct {
	Schema string  `yaml:"schema" json:"schema"`
	Actors []Actor `yaml:"actors" json:"actors"`
}

// Get returns the actor with the given id, or (nil, false).
func (r *Registry) Get(id string) (*Actor, bool) {
	for i := range r.Actors {
		if r.Actors[i].ID == id {
			return &r.Actors[i], true
		}
	}
	return nil, false
}

// Parse decodes registry YAML and validates it, returning the registry and any
// non-fatal warnings (e.g. unknown top-level keys, schema mismatch).
func Parse(data []byte) (*Registry, []string, error) {
	var reg Registry
	if err := yaml.Unmarshal(data, &reg); err != nil {
		return nil, nil, fmt.Errorf("parsing actors: %w", err)
	}

	var warnings []string
	if reg.Schema != SchemaVersion {
		warnings = append(warnings, fmt.Sprintf("actors schema %q does not match expected %q", reg.Schema, SchemaVersion))
	}

	if err := reg.validate(); err != nil {
		return nil, warnings, err
	}
	return &reg, warnings, nil
}

// Load reads and parses the registry at path.
func Load(path string) (*Registry, []string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, nil, err
	}
	return Parse(data)
}

// validate enforces the hard invariants: non-empty unique ids and known types.
func (r *Registry) validate() error {
	if len(r.Actors) == 0 {
		return fmt.Errorf("actor registry defines no actors")
	}
	seen := map[string]bool{}
	for i, a := range r.Actors {
		if a.ID == "" {
			return fmt.Errorf("actor #%d has an empty id", i+1)
		}
		if seen[a.ID] {
			return fmt.Errorf("duplicate actor id %q", a.ID)
		}
		seen[a.ID] = true
		if !a.Type.valid() {
			return fmt.Errorf("actor %q has invalid type %q (want human, ai_agent, ai_judge, or tool)", a.ID, a.Type)
		}
	}
	return nil
}

// IDs returns the actor ids, sorted, for stable display.
func (r *Registry) IDs() []string {
	ids := make([]string, len(r.Actors))
	for i, a := range r.Actors {
		ids[i] = a.ID
	}
	sort.Strings(ids)
	return ids
}
