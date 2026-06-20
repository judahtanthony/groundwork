package exporter

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
	"groundwork/internal/ticket"
)

// Parse is the inverse of Render: it reconstructs a work node and its dependency
// ids from exported Markdown (docs/contracts/ticket-export.md), so a cold store
// can be rebuilt from committed exports (T-0902). It tolerates the fallback
// placeholders Render emits for empty sections.
func Parse(data []byte) (*ticket.Ticket, []string, error) {
	s := strings.ReplaceAll(string(data), "\r\n", "\n")
	if !strings.HasPrefix(s, "---\n") {
		return nil, nil, fmt.Errorf("missing front matter opener")
	}
	rest := s[len("---\n"):]
	end := strings.Index(rest, "\n---\n")
	if end < 0 {
		return nil, nil, fmt.Errorf("missing front matter closer")
	}
	fmText := rest[:end+1]
	body := rest[end+len("\n---\n"):]

	var fm frontMatter
	if err := yaml.Unmarshal([]byte(fmText), &fm); err != nil {
		return nil, nil, fmt.Errorf("parsing front matter: %w", err)
	}
	if fm.ID == "" {
		return nil, nil, fmt.Errorf("export has no id")
	}

	t := &ticket.Ticket{
		ID:        fm.ID,
		Kind:      fm.Kind,
		Title:     fm.Title,
		Status:    ticket.Status(fm.Status),
		Priority:  fm.Priority,
		Labels:    fm.Labels,
		CreatedAt: fm.CreatedAt,
		UpdatedAt: fm.UpdatedAt,
	}
	if fm.NodeType != nil {
		t.NodeType = ticket.NodeType(*fm.NodeType)
	}
	if fm.WorkType != nil {
		t.WorkType = *fm.WorkType
	}
	if fm.Assignee != nil {
		t.Assignee = *fm.Assignee
	}
	if fm.RequestedActor != nil {
		t.RequestedActor = *fm.RequestedActor
	}
	if fm.Parent != nil {
		t.ParentID = *fm.Parent
	}

	sections := splitSections(body)
	if desc := strings.TrimSpace(sections["Problem"]); desc != "" && desc != "_No description recorded._" {
		t.Description = desc
	}
	t.Acceptance = parseBullets(sections["Acceptance Criteria"])
	if c := parseContract(sections["Design / Contract"]); c != "" {
		t.Contract = c
	}

	return t, fm.DependsOn, nil
}

// splitSections maps each "## Heading" to its content (until the next heading).
func splitSections(body string) map[string]string {
	out := map[string]string{}
	var heading string
	var buf []string
	flush := func() {
		if heading != "" {
			out[heading] = strings.TrimRight(strings.Join(buf, "\n"), "\n")
		}
		buf = nil
	}
	for _, line := range strings.Split(body, "\n") {
		if strings.HasPrefix(line, "## ") {
			flush()
			heading = strings.TrimSpace(strings.TrimPrefix(line, "## "))
			continue
		}
		buf = append(buf, line)
	}
	flush()
	return out
}

// parseBullets extracts "- item" lines, tolerating the "_None recorded._"
// placeholder.
func parseBullets(section string) []string {
	var out []string
	for _, line := range strings.Split(section, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "- ") {
			out = append(out, strings.TrimSpace(strings.TrimPrefix(line, "- ")))
		}
	}
	return out
}

// parseContract extracts the JSON inside a ```json fenced block, or "" if none.
func parseContract(section string) string {
	start := strings.Index(section, "```json\n")
	if start < 0 {
		return ""
	}
	rest := section[start+len("```json\n"):]
	end := strings.Index(rest, "\n```")
	if end < 0 {
		return ""
	}
	return strings.TrimSpace(rest[:end])
}
