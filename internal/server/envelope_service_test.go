package server

import (
	"strings"
	"testing"

	"groundwork/internal/approval"
	"groundwork/internal/envelope"
	"groundwork/internal/ticket"
)

// The approvals inbox shows the proposed boundary for an approve_envelope
// approval so the human sees what they would authorize (ADR 0054/T-1078).
func TestApprovalsInboxShowsEnvelopeBoundary(t *testing.T) {
	srv, db := newTestServer(t)
	parent := &ticket.Ticket{Title: "root", NodeType: ticket.NodeComposite, Status: ticket.StatusTodo, WorkType: "technical_design"}
	if err := db.CreateTicket(parent, "tester"); err != nil {
		t.Fatal(err)
	}
	draft := &envelope.Envelope{
		ApprovedActions: []string{envelope.ActionExecuteChildren},
		AllowedRoles:    []string{"coding"},
		RiskCeiling:     "medium",
		Planning:        envelope.Planning{AllowedWorkTypes: []string{"technical_implementation"}},
	}
	if _, err := srv.ProposeEnvelope(parent.ID, draft); err != nil {
		t.Fatal(err)
	}
	body := getHTML(t, srv, "/approvals").Body.String()
	for _, want := range []string{"approve_envelope", "envelope boundary:", "execute_children", "coding"} {
		if !strings.Contains(body, want) {
			t.Errorf("inbox missing %q", want)
		}
	}
}

// An envelope draft granting an unknown approved action is refused at propose
// time, so the human is never asked to approve a boundary that cannot activate
// (M5/ADR 0054).
func TestProposeEnvelopeRejectsUnknownAction(t *testing.T) {
	srv, db := newTestServer(t)
	parent := &ticket.Ticket{Title: "root", NodeType: ticket.NodeComposite, Status: ticket.StatusTodo, WorkType: "technical_design"}
	if err := db.CreateTicket(parent, "tester"); err != nil {
		t.Fatal(err)
	}
	if _, err := srv.ProposeEnvelope(parent.ID, &envelope.Envelope{
		ApprovedActions: []string{"land_to_main"}, // not an envelope-grantable action
		AllowedRoles:    []string{"coding"},
	}); err == nil {
		t.Fatal("ProposeEnvelope accepted an unknown approved action; want refusal")
	}
	if _, err := srv.ProposeEnvelope(parent.ID, &envelope.Envelope{AllowedRoles: []string{"coding"}}); err == nil {
		t.Fatal("ProposeEnvelope accepted an empty action set; want refusal")
	}
}

// Re-approving an already-activated envelope (a retried activation) does not
// allocate a second envelope: activation is idempotent (M2/ADR 0054).
func TestActivateEnvelopeIsIdempotent(t *testing.T) {
	srv, db := newTestServer(t)
	parent := &ticket.Ticket{Title: "root", NodeType: ticket.NodeComposite, Status: ticket.StatusTodo, WorkType: "technical_design"}
	if err := db.CreateTicket(parent, "tester"); err != nil {
		t.Fatal(err)
	}
	draft := &envelope.Envelope{ApprovedActions: []string{envelope.ActionExecuteChildren}, AllowedRoles: []string{"coding"}}
	appr, err := srv.ProposeEnvelope(parent.ID, draft)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := srv.recordDecision(appr.ID, approval.StatusApproved, ""); err != nil {
		t.Fatal(err)
	}
	first, _ := db.GetActiveEnvelopeForNode(parent.ID)
	if first == nil {
		t.Fatal("no active envelope after approval")
	}
	// A retried activation (same approval action JSON) must be a no-op, not a
	// duplicate envelope.
	if err := srv.activateEnvelope(appr.ActionJSON, parent.ID, ownerActor); err != nil {
		t.Fatalf("retried activation errored: %v", err)
	}
	again, _ := db.GetActiveEnvelopeForNode(parent.ID)
	if again == nil || again.ID != first.ID {
		t.Errorf("activation not idempotent: first=%v again=%v", first, again)
	}
}

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
