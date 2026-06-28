// Package completion is the child completion-summary domain (ADR 0047/0057): the
// compact record a node emits when it reaches review/landing, summarizing the
// outcome, changed files, validation, decisions, assumptions, and risks. It is
// file-authoritative — a per-node sidecar (.groundwork/tickets/<id>/completion.yaml)
// mirrored into SQLite — and is the unit the bulk review bundle aggregates.
package completion

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// ValidationLine records one validation command and its status.
type ValidationLine struct {
	Command string `yaml:"command" json:"command"`
	Status  string `yaml:"status" json:"status"`
}

// Summary is a node's completion record (ADR 0047).
type Summary struct {
	NodeID       string           `yaml:"node_id" json:"node_id"`
	Outcome      string           `yaml:"outcome" json:"outcome"`
	Changed      []string         `yaml:"changed" json:"changed,omitempty"`
	Validation   []ValidationLine `yaml:"validation" json:"validation,omitempty"`
	Decisions    []string         `yaml:"decisions" json:"decisions,omitempty"`
	Assumptions  []string         `yaml:"assumptions" json:"assumptions,omitempty"`
	Risks        []string         `yaml:"risks" json:"risks,omitempty"`
	CanonUpdates []string         `yaml:"canon_updates" json:"canon_updates,omitempty"`
	CreatedAt    string           `yaml:"created_at" json:"created_at,omitempty"`
}

// Stale reports whether a summary no longer reflects the node's current state
// (ADR 0047): the node was sent back to rework after the summary, or the run's
// changed-file set diverged from what the summary recorded. The reason is "" when
// the summary is current. A nil summary is not stale (there is nothing to age).
func Stale(s *Summary, currentChanged []string, status string) (bool, string) {
	if s == nil {
		return false, ""
	}
	if status == "rework" {
		return true, "node returned to rework after the summary was written"
	}
	if !equalSet(s.Changed, currentChanged) {
		return true, "changed-file set diverged from the summary"
	}
	return false, ""
}

// equalSet compares two file lists as sets (order-independent).
func equalSet(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	seen := make(map[string]int, len(a))
	for _, x := range a {
		seen[x]++
	}
	for _, y := range b {
		seen[y]--
		if seen[y] < 0 {
			return false
		}
	}
	return true
}

// SidecarPath returns the completion-summary sidecar path for a node.
func SidecarPath(ticketsDir, nodeID string) string {
	return filepath.Join(ticketsDir, nodeID, "completion.yaml")
}

// Write persists the summary as its node's sidecar (the authoritative copy).
func Write(ticketsDir string, s *Summary) error {
	dir := filepath.Join(ticketsDir, s.NodeID)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	data, err := yaml.Marshal(s)
	if err != nil {
		return err
	}
	return os.WriteFile(SidecarPath(ticketsDir, s.NodeID), data, 0o644)
}

// Read loads a node's completion sidecar. The bool is false when none exists.
func Read(ticketsDir, nodeID string) (*Summary, bool, error) {
	data, err := os.ReadFile(SidecarPath(ticketsDir, nodeID))
	if os.IsNotExist(err) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	var s Summary
	if err := yaml.Unmarshal(data, &s); err != nil {
		return nil, false, err
	}
	return &s, true, nil
}
