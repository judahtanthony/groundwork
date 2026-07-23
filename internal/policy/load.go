package policy

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"gopkg.in/yaml.v3"
)

// Set is the loaded policy bundle. Any field may be nil if its policy file is
// absent; the gate engine treats a nil policy conservatively (default-deny for
// claims, human-required for gated actions).
type Set struct {
	mu         sync.RWMutex
	Trust      *TrustPolicy
	Autonomy   *AutonomyPolicy
	Validation *ValidationPolicy
}

// ReplaceTrust swaps the live trust policy after a validated, gated policy
// amendment. Gate evaluation takes the same lock, so schedulers never observe a
// partially replaced rule set.
func (s *Set) ReplaceTrust(trust *TrustPolicy) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Trust = trust
}

// Load reads every *.yaml under dir, dispatches each by its `schema` field, and
// returns the combined Set plus warnings. A missing directory is not an error
// (an unconfigured project): it yields an empty Set and a warning. Two files of
// the same schema are an error to avoid ambiguous precedence.
func Load(dir string) (*Set, []string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return &Set{}, []string{fmt.Sprintf("no policies directory at %s", dir)}, nil
		}
		return nil, nil, err
	}

	var names []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".yaml") {
			names = append(names, e.Name())
		}
	}
	sort.Strings(names) // deterministic load order

	set := &Set{}
	var warnings []string
	for _, name := range names {
		path := filepath.Join(dir, name)
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, nil, err
		}
		ws, err := set.loadFile(name, data)
		if err != nil {
			return nil, nil, fmt.Errorf("%s: %w", name, err)
		}
		warnings = append(warnings, prefixWarnings(name, ws)...)
	}
	return set, warnings, nil
}

// loadFile dispatches one policy file by its declared schema.
func (s *Set) loadFile(name string, data []byte) ([]string, error) {
	var head struct {
		Schema string `yaml:"schema"`
	}
	if err := yaml.Unmarshal(data, &head); err != nil {
		return nil, fmt.Errorf("reading schema: %w", err)
	}
	switch head.Schema {
	case TrustSchema:
		if s.Trust != nil {
			return nil, fmt.Errorf("a trust policy is already loaded")
		}
		p, ws, err := ParseTrust(data)
		if err != nil {
			return ws, err
		}
		s.Trust = p
		return ws, nil
	case AutonomySchema:
		if s.Autonomy != nil {
			return nil, fmt.Errorf("an autonomy policy is already loaded")
		}
		p, ws, err := ParseAutonomy(data)
		if err != nil {
			return ws, err
		}
		s.Autonomy = p
		return ws, nil
	case ValidationSchema:
		if s.Validation != nil {
			return nil, fmt.Errorf("a validation policy is already loaded")
		}
		p, ws, err := ParseValidation(data)
		if err != nil {
			return ws, err
		}
		s.Validation = p
		return ws, nil
	default:
		return []string{fmt.Sprintf("unrecognized policy schema %q (ignored)", head.Schema)}, nil
	}
}

func prefixWarnings(name string, ws []string) []string {
	out := make([]string, len(ws))
	for i, w := range ws {
		out[i] = name + ": " + w
	}
	return out
}
