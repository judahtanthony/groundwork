package exporter

import (
	"bytes"
	"strings"
	"testing"

	"groundwork/internal/ticket"
)

func sampleLeaf() *ticket.Ticket {
	p := 0.5
	return &ticket.Ticket{
		ID:          "T-0001",
		Kind:        "ticket",
		NodeType:    ticket.NodeLeaf,
		Title:       "Implement SQLite migration runner",
		Description: "Groundwork needs a migration runner.",
		Status:      ticket.StatusTodo,
		Priority:    &p,
		Labels:      []string{"store", "sqlite"},
		Acceptance:  []string{"Migrations apply in order.", "Re-running is safe."},
		CreatedAt:   "2026-06-17T10:00:00Z",
		UpdatedAt:   "2026-06-17T10:00:00Z",
	}
}

func TestRenderIsDeterministic(t *testing.T) {
	tk := sampleLeaf()
	a, err := Render(tk, []string{"T-0000"})
	if err != nil {
		t.Fatal(err)
	}
	b, err := Render(tk, []string{"T-0000"})
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(a, b) {
		t.Fatal("Render is not byte-identical across calls")
	}
}

// TestRenderSortsDependsOn guards the determinism contract (ADR 0021): the
// rendered depends_on list must be sorted regardless of the order ids are
// supplied, so re-exporting an unchanged store yields a byte-identical tree.
func TestRenderSortsDependsOn(t *testing.T) {
	unsorted, err := Render(sampleLeaf(), []string{"T-1002", "T-0503", "T-0504"})
	if err != nil {
		t.Fatal(err)
	}
	sorted, err := Render(sampleLeaf(), []string{"T-0503", "T-0504", "T-1002"})
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(unsorted, sorted) {
		t.Fatalf("depends_on order leaks into output; export is not order-independent:\n%s", unsorted)
	}
	want := "depends_on:\n    - T-0503\n    - T-0504\n    - T-1002\n"
	if !strings.Contains(string(unsorted), want) {
		t.Errorf("depends_on not sorted; want %q in:\n%s", want, unsorted)
	}
}

// TestRenderPreservesEmptyTimestamps guards the committed empty-timestamp
// convention used by planning-sourced tickets: an empty created_at/updated_at
// must render as empty (not stamped), so a store rebuilt from such exports
// round-trips without churn (ADR 0021).
func TestRenderPreservesEmptyTimestamps(t *testing.T) {
	tk := sampleLeaf()
	tk.CreatedAt = ""
	tk.UpdatedAt = ""
	out, err := Render(tk, nil)
	if err != nil {
		t.Fatal(err)
	}
	s := string(out)
	for _, want := range []string{`created_at: ""`, `updated_at: ""`} {
		if !strings.Contains(s, want) {
			t.Errorf("expected %q in:\n%s", want, s)
		}
	}
}

func TestRenderContent(t *testing.T) {
	out, err := Render(sampleLeaf(), []string{"T-0000"})
	if err != nil {
		t.Fatal(err)
	}
	s := string(out)
	for _, want := range []string{
		"---\n",
		"id: T-0001",
		"node_type: leaf",
		"work_type: null",
		"requested_actor: null",
		"depends_on:\n    - T-0000",
		"## Problem",
		"## Acceptance Criteria",
		"- Migrations apply in order.",
	} {
		if !strings.Contains(s, want) {
			t.Errorf("export missing %q:\n%s", want, s)
		}
	}
	if !strings.HasSuffix(s, "\n") || strings.HasSuffix(s, "\n\n") {
		t.Errorf("export must end with exactly one trailing newline")
	}
}

func TestRenderNullsForEmptyFields(t *testing.T) {
	tk := &ticket.Ticket{
		ID: "T-0009", Kind: "ticket", Title: "untriaged", Status: ticket.StatusBacklog,
		CreatedAt: "2026-06-17T10:00:00Z", UpdatedAt: "2026-06-17T10:00:00Z",
	}
	out, _ := Render(tk, nil)
	s := string(out)
	for _, want := range []string{"node_type: null", "assignee: null", "parent: null", "depends_on: []", "labels: []"} {
		if !strings.Contains(s, want) {
			t.Errorf("expected %q in:\n%s", want, s)
		}
	}
}

func TestRenderActorFields(t *testing.T) {
	tk := sampleLeaf()
	tk.WorkType = "technical_implementation"
	tk.RequestedActor = "ai.codex.default"
	out, _ := Render(tk, nil)
	s := string(out)
	for _, want := range []string{"work_type: technical_implementation", "requested_actor: ai.codex.default"} {
		if !strings.Contains(s, want) {
			t.Errorf("export missing %q:\n%s", want, s)
		}
	}
}

func TestRenderCompositeSections(t *testing.T) {
	tk := sampleLeaf()
	tk.NodeType = ticket.NodeComposite
	out, _ := Render(tk, nil)
	s := string(out)
	if !strings.Contains(s, "## Design / Contract") || !strings.Contains(s, "## Escalations") {
		t.Errorf("composite export missing sections:\n%s", s)
	}
}
