package server

import (
	"testing"

	"groundwork/internal/envelope"
	"groundwork/internal/risk"
	"groundwork/internal/ticket"
)

func TestEnvelopeFacts(t *testing.T) {
	srv, db := newTestServer(t)
	root := &ticket.Ticket{Title: "root", NodeType: ticket.NodeComposite, Status: ticket.StatusTodo, WorkType: "technical_design"}
	if err := db.CreateTicket(root, "tester"); err != nil {
		t.Fatal(err)
	}
	child := &ticket.Ticket{ParentID: root.ID, Title: "child", NodeType: ticket.NodeLeaf, Status: ticket.StatusInProgress, WorkType: "technical_implementation"}
	if err := db.CreateTicket(child, "tester"); err != nil {
		t.Fatal(err)
	}
	env := &envelope.Envelope{
		ID: "ENV-0001", NodeID: root.ID, Status: envelope.StatusActive,
		ApprovedActions: []string{envelope.ActionExecuteChildren},
		AllowedRoles:    []string{"coding"},
		Planning:        envelope.Planning{AllowedWorkTypes: []string{"technical_implementation"}},
		RiskCeiling:     "medium",
		Scope:           envelope.Scope{Files: envelope.FileScope{Allow: []string{"internal/**"}, Deny: []string{"**/*secret*"}}},
	}
	if err := db.UpsertEnvelope(env); err != nil {
		t.Fatal(err)
	}

	act := envelope.ActionExecuteChildren
	// In scope: action/role/work-type/risk/files all within the root envelope.
	id, within, planned, err := srv.envelopeFacts(child.ID, act, "coding", "technical_implementation", risk.ClassLow, []string{"internal/x.go"})
	if err != nil {
		t.Fatal(err)
	}
	if id != "ENV-0001" || !within {
		t.Errorf("in-scope: id=%q within=%v, want ENV-0001/true", id, within)
	}
	if len(planned) != 1 || planned[0] != "internal/**" {
		t.Errorf("planned scope = %v", planned)
	}

	// Out of file scope.
	if _, within, _, _ := srv.envelopeFacts(child.ID, act, "coding", "technical_implementation", risk.ClassLow, []string{"cmd/main.go"}); within {
		t.Error("out-of-scope file should not be within envelope")
	}
	// Denied file.
	if _, within, _, _ := srv.envelopeFacts(child.ID, act, "coding", "technical_implementation", risk.ClassLow, []string{"internal/secret.go"}); within {
		t.Error("denied file should not be within envelope")
	}
	// Wrong role.
	if _, within, _, _ := srv.envelopeFacts(child.ID, act, "planner", "technical_implementation", risk.ClassLow, []string{"internal/x.go"}); within {
		t.Error("disallowed role should not be within envelope")
	}
	// Risk above ceiling.
	if _, within, _, _ := srv.envelopeFacts(child.ID, act, "coding", "technical_implementation", risk.ClassCritical, []string{"internal/x.go"}); within {
		t.Error("risk above ceiling should not be within envelope")
	}
	// Action not approved.
	if _, within, _, _ := srv.envelopeFacts(child.ID, envelope.ActionDecomposeChildren, "coding", "technical_implementation", risk.ClassLow, []string{"internal/x.go"}); within {
		t.Error("unapproved action should not be within envelope")
	}
	// No envelope in chain.
	orphan := &ticket.Ticket{Title: "orphan", NodeType: ticket.NodeLeaf, Status: ticket.StatusInProgress, WorkType: "technical_implementation"}
	if err := db.CreateTicket(orphan, "tester"); err != nil {
		t.Fatal(err)
	}
	if id, within, _, _ := srv.envelopeFacts(orphan.ID, act, "coding", "technical_implementation", risk.ClassLow, nil); id != "" || within {
		t.Errorf("orphan: id=%q within=%v, want empty/false", id, within)
	}
}
