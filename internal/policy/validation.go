package policy

import (
	"sort"

	"groundwork/internal/risk"
)

// RequiredChecks returns the validation checks required for the given changed
// files, deduplicated by name across all matching templates (T-0701). The result
// is sorted by name for determinism. A nil policy requires nothing.
func (vp *ValidationPolicy) RequiredChecks(files []string) []ValidationCheck {
	if vp == nil {
		return nil
	}
	seen := map[string]ValidationCheck{}
	for _, tmpl := range vp.Templates {
		if anyFileMatch(tmpl.Match.Files, files) {
			for _, c := range tmpl.Required {
				if _, ok := seen[c.Name]; !ok {
					seen[c.Name] = c
				}
			}
		}
	}
	out := make([]ValidationCheck, 0, len(seen))
	for _, c := range seen {
		out = append(out, c)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

// RequiredCheckNames returns just the names of the required checks for files.
func (vp *ValidationPolicy) RequiredCheckNames(files []string) []string {
	checks := vp.RequiredChecks(files)
	names := make([]string, len(checks))
	for i, c := range checks {
		names[i] = c.Name
	}
	return names
}

// LandingRiskFloor returns the highest landing risk floor among templates
// matching files, or empty if none set. The floor raises a node's risk class at
// landing regardless of its computed score (validation.md).
func (vp *ValidationPolicy) LandingRiskFloor(files []string) risk.Class {
	if vp == nil {
		return ""
	}
	floor := risk.Class("")
	for _, tmpl := range vp.Templates {
		if tmpl.LandingRiskFloor == "" || !anyFileMatch(tmpl.Match.Files, files) {
			continue
		}
		c := risk.Class(tmpl.LandingRiskFloor)
		if floor == "" || floor.AtMost(c) {
			floor = c
		}
	}
	return floor
}
