package sqlite

import (
	"testing"

	"groundwork/internal/envelope"
	"groundwork/internal/ticket"
)

func TestEnvelopeMirrorCRUD(t *testing.T) {
	db := openTestDB(t)
	parent := &ticket.Ticket{Title: "root", NodeType: ticket.NodeComposite, Status: ticket.StatusTodo, WorkType: "technical_design"}
	if err := db.CreateTicket(parent, "tester"); err != nil {
		t.Fatal(err)
	}
	e := &envelope.Envelope{
		ID: "env-" + parent.ID, NodeID: parent.ID, Status: envelope.StatusActive,
		ApprovedBy: "human.owner", ApprovedActions: []string{envelope.ActionExecuteChildren},
		AllowedRoles: []string{"coding"},
	}
	if err := db.UpsertEnvelope(e); err != nil {
		t.Fatal(err)
	}

	got, err := db.GetActiveEnvelopeForNode(parent.ID)
	if err != nil || got == nil {
		t.Fatalf("active envelope: got=%v err=%v", got, err)
	}
	if got.ID != e.ID || !got.Allows(envelope.ActionExecuteChildren) {
		t.Errorf("mirror mismatch: %+v", got)
	}

	// Revoke -> no active envelope for the node.
	if err := db.SetEnvelopeStatus(e.ID, envelope.StatusRevoked); err != nil {
		t.Fatal(err)
	}
	if got, _ := db.GetActiveEnvelopeForNode(parent.ID); got != nil {
		t.Errorf("active envelope after revoke = %v, want nil", got)
	}
	if byID, err := db.GetEnvelope(e.ID); err != nil || byID.Status != envelope.StatusRevoked {
		t.Errorf("GetEnvelope after revoke: %+v err=%v", byID, err)
	}
}
