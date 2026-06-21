// Package exporter renders work nodes to deterministic Markdown for durable,
// committed state (docs/contracts/ticket-export.md, ADR 0020). Rendering is pure
// and byte-stable: equal inputs always produce identical output.
package exporter

import (
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
	"groundwork/internal/ticket"
)

// frontMatter is the exported YAML front matter. Field order is the canonical
// key order (yaml.v3 marshals struct fields in declaration order, ADR 0020).
// Pointer fields render as `null` when nil.
type frontMatter struct {
	ID             string   `yaml:"id"`
	Kind           string   `yaml:"kind"`
	NodeType       *string  `yaml:"node_type"`
	WorkType       *string  `yaml:"work_type"`
	Title          string   `yaml:"title"`
	Status         string   `yaml:"status"`
	Assignee       *string  `yaml:"assignee"`
	RequestedActor *string  `yaml:"requested_actor"`
	Priority       *float64 `yaml:"priority"`
	Labels         []string `yaml:"labels"`
	Parent         *string  `yaml:"parent"`
	DependsOn      []string `yaml:"depends_on"`
	CreatedAt      string   `yaml:"created_at"`
	UpdatedAt      string   `yaml:"updated_at"`
}

// Render returns the deterministic Markdown export of t, with dependsOn as the
// node's dependency ids (already sorted by the caller). Output uses LF line
// endings and a single trailing newline.
func Render(t *ticket.Ticket, dependsOn []string) ([]byte, error) {
	fm := frontMatter{
		ID:             t.ID,
		Kind:           t.Kind,
		NodeType:       ptrOrNil(string(t.NodeType)),
		WorkType:       ptrOrNil(t.WorkType),
		Title:          t.Title,
		Status:         string(t.Status),
		Assignee:       ptrOrNil(t.Assignee),
		RequestedActor: ptrOrNil(t.RequestedActor),
		Priority:       t.Priority,
		Labels:         nonNil(t.Labels),
		Parent:         ptrOrNil(t.ParentID),
		DependsOn:      nonNil(dependsOn),
		CreatedAt:      t.CreatedAt,
		UpdatedAt:      t.UpdatedAt,
	}
	yamlBytes, err := yaml.Marshal(fm)
	if err != nil {
		return nil, err
	}

	var b strings.Builder
	b.WriteString("---\n")
	b.Write(yamlBytes)
	b.WriteString("---\n\n")

	b.WriteString("## Problem\n\n")
	b.WriteString(bodyText(t.Description, "_No description recorded._"))

	b.WriteString("\n## Acceptance Criteria\n\n")
	if len(t.Acceptance) == 0 {
		b.WriteString("_None recorded._\n")
	} else {
		for _, a := range t.Acceptance {
			b.WriteString("- ")
			b.WriteString(a)
			b.WriteString("\n")
		}
	}

	// Composite nodes carry a Design/Contract section and an Escalations section
	// (docs/contracts/ticket-export.md). Escalation content is Phase 2.
	if t.NodeType == ticket.NodeComposite {
		b.WriteString("\n## Design / Contract\n\n")
		if t.Contract == "" || t.Contract == "{}" {
			b.WriteString("_No contract recorded._\n")
		} else {
			b.WriteString("```json\n")
			b.WriteString(t.Contract)
			b.WriteString("\n```\n")
		}

		b.WriteString("\n## Escalations\n\n")
		b.WriteString("_No escalations._\n")
	}

	return []byte(b.String()), nil
}

// WriteTo renders t and writes ticketsDir/<id>/ticket.md, returning the path.
// It is the shared writer for `gw ticket export` and the server's landing
// commit, so both produce byte-identical exports.
func WriteTo(ticketsDir string, t *ticket.Ticket, dependsOn []string) (string, error) {
	data, err := Render(t, dependsOn)
	if err != nil {
		return "", err
	}
	dir := filepath.Join(ticketsDir, t.ID)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	path := filepath.Join(dir, "ticket.md")
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return "", err
	}
	return path, nil
}

func bodyText(s, fallback string) string {
	s = strings.TrimRight(s, "\n")
	if strings.TrimSpace(s) == "" {
		return fallback + "\n"
	}
	return s + "\n"
}

func ptrOrNil(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func nonNil(in []string) []string {
	if in == nil {
		return []string{}
	}
	return in
}
