package exporter

import (
	"reflect"
	"testing"

	"groundwork/internal/ticket"
)

func TestRenderParseRoundTrip(t *testing.T) {
	prio := 0.5
	orig := &ticket.Ticket{
		ID:             "T-0007",
		Kind:           "epic",
		NodeType:       ticket.NodeComposite,
		WorkType:       "technical_design",
		Title:          "Build the thing",
		Description:    "A multi-line\n\ndescription with detail.",
		Contract:       `{"schema":"contract/v1"}`,
		Status:         ticket.StatusTodo,
		Assignee:       "human.owner",
		RequestedActor: "ai.codex.default",
		Priority:       &prio,
		Labels:         []string{"store", "sqlite"},
		Acceptance:     []string{"does X", "does Y"},
		ParentID:       "G-0001",
		CreatedAt:      "2026-06-17T10:00:00Z",
		UpdatedAt:      "2026-06-18T10:00:00Z",
	}
	deps := []string{"T-0006"}

	data, err := Render(orig, deps)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	got, gotDeps, err := Parse(data)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	if got.ID != orig.ID || got.Kind != orig.Kind || got.Title != orig.Title ||
		got.Status != orig.Status || got.NodeType != orig.NodeType || got.WorkType != orig.WorkType ||
		got.Assignee != orig.Assignee || got.RequestedActor != orig.RequestedActor || got.ParentID != orig.ParentID {
		t.Fatalf("scalar mismatch:\n got=%+v\nwant=%+v", got, orig)
	}
	if got.Description != orig.Description {
		t.Errorf("description = %q, want %q", got.Description, orig.Description)
	}
	if !reflect.DeepEqual(got.Acceptance, orig.Acceptance) {
		t.Errorf("acceptance = %v, want %v", got.Acceptance, orig.Acceptance)
	}
	if !reflect.DeepEqual(got.Labels, orig.Labels) {
		t.Errorf("labels = %v, want %v", got.Labels, orig.Labels)
	}
	if got.Contract != orig.Contract {
		t.Errorf("contract = %q, want %q", got.Contract, orig.Contract)
	}
	if got.Priority == nil || *got.Priority != prio {
		t.Errorf("priority = %v, want %g", got.Priority, prio)
	}
	if !reflect.DeepEqual(gotDeps, deps) {
		t.Errorf("deps = %v, want %v", gotDeps, deps)
	}
	if got.CreatedAt != orig.CreatedAt || got.UpdatedAt != orig.UpdatedAt {
		t.Errorf("timestamps not preserved: %s / %s", got.CreatedAt, got.UpdatedAt)
	}
}

func TestParseLeafNoOptionalSections(t *testing.T) {
	orig := &ticket.Ticket{ID: "T-0001", Kind: "ticket", Title: "Leaf", Status: ticket.StatusTodo}
	data, err := Render(orig, nil)
	if err != nil {
		t.Fatal(err)
	}
	got, deps, err := Parse(data)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if got.Description != "" || len(got.Acceptance) != 0 || got.Contract != "" || len(deps) != 0 {
		t.Errorf("leaf parse should be empty optionals: %+v deps=%v", got, deps)
	}
}
