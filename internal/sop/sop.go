// Package sop loads work-type standard operating procedures from
// .groundwork/sops/<work-type>/ (ADR 0011). SOPs are committed canon that
// supplies a planning/execution agent its work-type instructions and context,
// and is one of the inputs that lets an action's autonomy level be elevated.
// Elevation itself is always a human act (trust-and-approvals.md); this package
// only reads SOPs.
package sop

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Dir returns the SOP directory for a work type under sopsDir, or "" if the work
// type is empty.
func Dir(sopsDir, workType string) string {
	if workType == "" {
		return ""
	}
	return filepath.Join(sopsDir, workType)
}

// List returns the SOP file paths for a work type, sorted, relative to sopsDir.
// A missing directory yields an empty list (not an error): most work types have
// no SOP yet.
func List(sopsDir, workType string) ([]string, error) {
	dir := Dir(sopsDir, workType)
	if dir == "" {
		return []string{}, nil
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, err
	}
	out := []string{}
	for _, e := range entries {
		if e.IsDir() || strings.HasPrefix(e.Name(), ".") {
			continue
		}
		out = append(out, filepath.Join(workType, e.Name()))
	}
	sort.Strings(out)
	return out, nil
}

// Load returns the SOP file contents for a work type, keyed by path relative to
// sopsDir. A missing directory yields an empty map.
func Load(sopsDir, workType string) (map[string]string, error) {
	rels, err := List(sopsDir, workType)
	if err != nil {
		return nil, err
	}
	out := map[string]string{}
	for _, rel := range rels {
		data, err := os.ReadFile(filepath.Join(sopsDir, rel))
		if err != nil {
			return nil, err
		}
		out[rel] = string(data)
	}
	return out, nil
}
