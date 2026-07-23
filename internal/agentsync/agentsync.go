// Package agentsync keeps a small, owned Groundwork instruction block in
// repository-level AGENTS.md files. It never rewrites instructions outside the
// managed markers.
package agentsync

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	startMarker = "<!-- GROUNDWORK_START -->"
	endMarker   = "<!-- GROUNDWORK_END -->"
)

const managedBlock = startMarker + `
## Groundwork

This repository uses Groundwork for its live work plan. Before changing the
project, read ` + "`.groundwork/WORKFLOW.md`" + ` and inspect the current tree with
` + "`gw ticket tree`" + `.
` + endMarker

// Status describes whether the repository's AGENTS.md contains the current
// Groundwork-managed instruction block.
type Status struct {
	Path   string `json:"path"`
	State  string `json:"state"`
	Detail string `json:"detail"`
}

// Inspect reports missing, out_of_sync, or synced without changing the file.
func Inspect(root string) (Status, error) {
	path := filepath.Join(root, "AGENTS.md")
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return Status{Path: path, State: "missing", Detail: "AGENTS.md does not exist; sync will create it."}, nil
	}
	if err != nil {
		return Status{}, err
	}
	text := string(data)
	start := strings.Index(text, startMarker)
	end := strings.Index(text, endMarker)
	if start < 0 || end < start {
		return Status{Path: path, State: "out_of_sync", Detail: "Groundwork's managed instruction block is missing."}, nil
	}
	end += len(endMarker)
	if text[start:end] != managedBlock {
		return Status{Path: path, State: "out_of_sync", Detail: "Groundwork's managed instruction block has changed."}, nil
	}
	return Status{Path: path, State: "synced", Detail: "Groundwork's managed instruction block is current."}, nil
}

// Sync creates or replaces only the managed Groundwork block.
func Sync(root string) (Status, error) {
	path := filepath.Join(root, "AGENTS.md")
	data, err := os.ReadFile(path)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return Status{}, err
	}
	text := string(data)
	start := strings.Index(text, startMarker)
	end := strings.Index(text, endMarker)
	switch {
	case start >= 0 && end >= start:
		end += len(endMarker)
		text = text[:start] + managedBlock + text[end:]
	case start >= 0 || end >= 0:
		return Status{}, fmt.Errorf("AGENTS.md contains an incomplete Groundwork managed block")
	default:
		if text != "" && !strings.HasSuffix(text, "\n") {
			text += "\n"
		}
		if text != "" {
			text += "\n"
		}
		text += managedBlock + "\n"
	}

	mode := os.FileMode(0o644)
	if info, statErr := os.Stat(path); statErr == nil {
		mode = info.Mode().Perm()
	}
	tmp, err := os.CreateTemp(root, ".agents-md-*")
	if err != nil {
		return Status{}, err
	}
	tmpPath := tmp.Name()
	defer os.Remove(tmpPath)
	if err := tmp.Chmod(mode); err != nil {
		tmp.Close()
		return Status{}, err
	}
	if _, err := tmp.WriteString(text); err != nil {
		tmp.Close()
		return Status{}, err
	}
	if err := tmp.Close(); err != nil {
		return Status{}, err
	}
	if err := os.Rename(tmpPath, path); err != nil {
		return Status{}, err
	}
	return Inspect(root)
}
