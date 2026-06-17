// Package config implements project-root discovery and parsing of
// .groundwork/config.yaml (see ADR 0021 and docs/contracts/file-layout.md).
package config

import (
	"fmt"
	"time"

	"gopkg.in/yaml.v3"
)

// SchemaVersion is the config schema this build understands.
const SchemaVersion = "groundwork_config/v1"

// Config is the parsed .groundwork/config.yaml. Phase 1 keys only; later phases
// add fields via new keys (unknown keys warn rather than error).
type Config struct {
	Schema         string `yaml:"schema"`
	Runtime        string `yaml:"runtime"`
	Server         Server `yaml:"server"`
	MaxConcurrency int    `yaml:"max_concurrency"`
	Lease          Lease  `yaml:"lease"`
	Sandbox        string `yaml:"sandbox"`
}

// Server holds coordinator network settings (used in Phase 2).
type Server struct {
	Addr string `yaml:"addr"`
}

// Lease holds claim-lease timing (docs/architecture/runtime-model.md).
type Lease struct {
	TTL       Duration `yaml:"ttl"`
	Heartbeat Duration `yaml:"heartbeat"`
}

// Defaults returns a Config populated with the Phase 1 default values.
func Defaults() Config {
	return Config{
		Schema:         SchemaVersion,
		Runtime:        "codex",
		Server:         Server{Addr: "127.0.0.1:4500"},
		MaxConcurrency: 4,
		Lease:          Lease{TTL: Duration(90 * time.Second), Heartbeat: Duration(30 * time.Second)},
		Sandbox:        "workspace-write",
	}
}

// knownKeys is the set of recognized top-level config keys, used to warn on
// unknown keys without rejecting forward-compatible files.
var knownKeys = map[string]bool{
	"schema":          true,
	"runtime":         true,
	"server":          true,
	"max_concurrency": true,
	"lease":           true,
	"sandbox":         true,
}

// Parse decodes config YAML over the Phase 1 defaults, returning the merged
// config and a list of warnings (e.g. unknown top-level keys).
func Parse(data []byte) (*Config, []string, error) {
	cfg := Defaults()

	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, nil, fmt.Errorf("parsing config: %w", err)
	}

	var warnings []string
	var top map[string]yaml.Node
	if err := yaml.Unmarshal(data, &top); err == nil {
		for k := range top {
			if !knownKeys[k] {
				warnings = append(warnings, fmt.Sprintf("unknown config key %q (ignored)", k))
			}
		}
	}

	if cfg.Schema != SchemaVersion {
		warnings = append(warnings, fmt.Sprintf("config schema %q does not match expected %q", cfg.Schema, SchemaVersion))
	}

	return &cfg, warnings, nil
}

// Marshal renders cfg as YAML suitable for writing config.yaml.
func Marshal(cfg *Config) ([]byte, error) {
	return yaml.Marshal(cfg)
}

// Duration is a time.Duration that (un)marshals as a Go duration string ("90s").
type Duration time.Duration

// Duration returns the underlying time.Duration.
func (d Duration) Duration() time.Duration { return time.Duration(d) }

// UnmarshalYAML parses a duration string such as "90s".
func (d *Duration) UnmarshalYAML(value *yaml.Node) error {
	var s string
	if err := value.Decode(&s); err != nil {
		return fmt.Errorf("duration must be a string like \"90s\": %w", err)
	}
	parsed, err := time.ParseDuration(s)
	if err != nil {
		return fmt.Errorf("invalid duration %q: %w", s, err)
	}
	*d = Duration(parsed)
	return nil
}

// MarshalYAML renders the duration as a Go duration string.
func (d Duration) MarshalYAML() (any, error) {
	return time.Duration(d).String(), nil
}
