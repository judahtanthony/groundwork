package server

import (
	"testing"

	"groundwork/internal/actor"
	"groundwork/internal/policy"
	"groundwork/internal/ticket"
)

// When a require_human rule names a required approver role, the opened approval
// records it (ADR 0055), so the decision is honest about which role was required.
func TestApprovalRecordsRequiredRoleFromPolicy(t *testing.T) {
	srv, db := newTestServer(t)
	reg, _, err := actor.Parse([]byte(testActorsYAML))
	if err != nil {
		t.Fatal(err)
	}
	pol := &policy.Set{Trust: &policy.TrustPolicy{
		RequireHuman: []policy.Rule{{
			ID: "land-needs-staff", When: policy.Match{ActionTypes: []string{"land_to_main"}},
			RequireRoles: []string{"staff_engineer"},
		}},
	}}
	srv.SetApprovals(NewApprovalService(db, pol, reg))

	tk := &ticket.Ticket{Title: "land me", Status: ticket.StatusReview, WorkType: "technical_implementation"}
	if err := db.CreateTicket(tk, "tester"); err != nil {
		t.Fatal(err)
	}
	appr, err := srv.approvals.RequestLanding(tk.ID, tk.WorkType)
	if err != nil {
		t.Fatal(err)
	}
	if len(appr.RequiredRoles) != 1 || appr.RequiredRoles[0] != "staff_engineer" {
		t.Errorf("approval required roles = %v, want [staff_engineer]", appr.RequiredRoles)
	}
}
