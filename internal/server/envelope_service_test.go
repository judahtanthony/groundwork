package server

import (
	"testing"

	"groundwork/internal/approval"
	"groundwork/internal/envelope"
	"groundwork/internal/ticket"
)

// Proposing an envelope opens a pending approve_envelope approval; approving it
// materializes an active envelope (mirror + authoritative sidecar). Revoking
// clears the active envelope. (ADR 0054)
func TestEnvelopeProposeApproveRevoke(t *testing.T) {
	srv, db := newTestServer(t)
	parent := &ticket.Ticket{Title: "root", NodeType: ticket.NodeComposite, Status: ticket.StatusTodo, WorkType: "technical_design"}
	if err := db.CreateTicket(parent, "tester"); err != nil {
		t.Fatal(err)
	}
	draft := &envelope.Envelope{
		ApprovedActions: []string{envelope.ActionExecuteChildren, envelope.ActionLandChildToParent},
		AllowedRoles:    []string{"coding"},
		RiskCeiling:     "medium",
		Planning:        envelope.Planning{MaxDepth: 2, AllowedWorkTypes: []string{"technical_implementation"}},
	}
	appr, err := srv.ProposeEnvelope(parent.ID, draft)
	if err != nil {
		t.Fatal(err)
	}
	if appr.Status != string(approval.StatusPending) || appr.Type != string(approval.TypeApproveEnvelope) {
		t.Fatalf("proposal = %s/%s, want pending/approve_envelope", appr.Status, appr.Type)
	}
	// No active envelope until approved.
	if got, _ := db.GetActiveEnvelopeForNode(parent.ID); got != nil {
		t.Fatal("envelope active before approval")
	}

	if _, err := srv.recordDecision(appr.ID, approval.StatusApproved, ""); err != nil {
		t.Fatal(err)
	}
	got, err := db.GetActiveEnvelopeForNode(parent.ID)
	if err != nil || got == nil {
		t.Fatalf("active envelope after approval: got=%v err=%v", got, err)
	}
	if got.Status != envelope.StatusActive || !got.Allows(envelope.ActionExecuteChildren) || got.ApprovedBy != ownerActor {
		t.Errorf("activated envelope wrong: %+v", got)
	}
	// Authoritative sidecar written.
	if e, ok, _ := envelope.Read(srv.proj.TicketsDir(), parent.ID); !ok || e.Status != envelope.StatusActive {
		t.Errorf("sidecar missing/wrong: ok=%v", ok)
	}

	if err := srv.RevokeEnvelope(got.ID); err != nil {
		t.Fatal(err)
	}
	if a, _ := db.GetActiveEnvelopeForNode(parent.ID); a != nil {
		t.Error("envelope still active after revoke")
	}
	if e, _, _ := envelope.Read(srv.proj.TicketsDir(), parent.ID); e == nil || e.Status != envelope.StatusRevoked {
		t.Errorf("sidecar not updated to revoked: %+v", e)
	}
}
