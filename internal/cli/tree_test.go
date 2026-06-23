package cli

import (
	"encoding/json"
	"strings"
	"testing"

	"groundwork/internal/ticket"
)

func TestBuildTreeNodesCarriesShowFields(t *testing.T) {
	prio := 0.7
	parent := &ticket.Ticket{ID: "T-0001", Title: "parent", Status: ticket.StatusTodo, NodeType: ticket.NodeComposite}
	child := &ticket.Ticket{
		ID: "T-0002", Title: "child", Status: ticket.StatusTodo, ParentID: "T-0001",
		WorkType: "technical_implementation", Priority: &prio,
	}
	children := map[string][]*ticket.Ticket{
		"":       {parent},
		"T-0001": {child},
	}

	nodes := buildTreeNodes(children[""], children)
	if len(nodes) != 1 || len(nodes[0].Children) != 1 {
		t.Fatalf("unexpected tree shape: %+v", nodes)
	}
	c := nodes[0].Children[0]
	if c.ParentID != "T-0001" || c.WorkType != "technical_implementation" {
		t.Errorf("child parent/work_type = %q/%q", c.ParentID, c.WorkType)
	}
	if c.Priority == nil || *c.Priority != 0.7 {
		t.Errorf("child priority = %v, want 0.7", c.Priority)
	}

	// The JSON now exposes priority/parent_id/work_type, so callers need not fall
	// back to `show --json` (ADR 0041).
	data, err := json.Marshal(nodes)
	if err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{`"priority"`, `"parent_id"`, `"work_type"`} {
		if !strings.Contains(string(data), key) {
			t.Errorf("tree JSON missing %s: %s", key, data)
		}
	}
}
