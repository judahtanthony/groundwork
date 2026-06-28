package server

import (
	"testing"

	"groundwork/internal/actor"
	"groundwork/internal/approval"
	"groundwork/internal/policy"
	"groundwork/internal/risk"
	"groundwork/internal/store/sqlite"
	"groundwork/internal/ticket"
)

// docsAutoApprovePolicy auto-approves docs-only changes but keeps landing human.
func docsAutoApprovePolicy() *policy.Set {
	return &policy.Set{Trust: &policy.TrustPolicy{
		AutoApprove: []policy.Rule{{
			ID:   "internal_docs",
			When: policy.Match{Files: []string{"**/*.md"}, ChangeType: "documentation"},
		}},
		RequireHuman: []policy.Rule{{
			ID:   "landing_to_main_v1",
			When: policy.Match{ActionTypes: []string{"land_to_main"}},
		}},
	}}
}

func requestService(t *testing.T) (*ApprovalService, *sqlite.DB) {
	t.Helper()
	db, err := sqlite.Open(t.TempDir() + "/state.sqlite")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	if err := db.Migrate(); err != nil {
		t.Fatal(err)
	}
	reg, _, err := actor.Parse([]byte(testActorsYAML))
	if err != nil {
		t.Fatal(err)
	}
	return NewApprovalService(db, docsAutoApprovePolicy(), reg), db
}

func TestRequestAutoApprovesDocs(t *testing.T) {
	svc, db := requestService(t)
	tk := &ticket.Ticket{Title: "docs", Status: ticket.StatusInProgress}
	if err := db.CreateTicket(tk, "tester"); err != nil {
		t.Fatal(err)
	}

	a, err := svc.Request(RequestParams{
		TicketID: tk.ID, Type: approval.TypeExecute, Summary: "edit docs",
		Action: policy.Action{Type: "execute", ChangeType: "documentation",
			Scope: risk.Scope{Files: []string{"docs/readme.md"}}},
	})
	if err != nil {
		t.Fatalf("Request: %v", err)
	}
	if a.Status != string(approval.StatusApproved) {
		t.Fatalf("status = %s, want approved", a.Status)
	}
	if a.DecidedByActor != "policy:internal_docs" {
		t.Errorf("decided_by = %q, want policy:internal_docs", a.DecidedByActor)
	}
}

// reviewerRoleActorsYAML gives an AI agent the reviewer role and a human a
// non-matching role, to probe the human-gate vs. required-role interaction (H1).
const reviewerRoleActorsYAML = `schema: groundwork_actors/v1
actors:
  - id: human.owner
    type: human
    display_name: Owner
    roles: [owner, reviewer]
  - id: ai.reviewer.codex
    type: ai_agent
    display_name: Reviewer Bot
    runtime: codex
    roles: [reviewer]
`

// A human-gated approval that also names a required role must NOT be decidable by
// an AI actor that merely holds that role: require_human is enforced first and
// independently of the role constraint (ADR 0028/0055; regression from T-1067
// which began auto-populating RequiredRoles from the firing rule). The required
// role only narrows *which human* may decide.
func TestHumanGatedApprovalRejectsAIWithRequiredRole(t *testing.T) {
	db, err := sqlite.Open(t.TempDir() + "/state.sqlite")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	if err := db.Migrate(); err != nil {
		t.Fatal(err)
	}
	reg, _, err := actor.Parse([]byte(reviewerRoleActorsYAML))
	if err != nil {
		t.Fatal(err)
	}
	svc := NewApprovalService(db, &policy.Set{}, reg)

	tk := &ticket.Ticket{Title: "boundary", Status: ticket.StatusInProgress}
	if err := db.CreateTicket(tk, "tester"); err != nil {
		t.Fatal(err)
	}
	score := 0
	rev := true
	a, err := db.CreateApproval(sqlite.CreateApprovalParams{
		TicketID: tk.ID, Type: approval.TypeApproveEnvelope, Status: approval.StatusPending,
		RiskClass: string(risk.ClassLow), RiskScore: &score, Reversible: &rev,
		Summary: "approve envelope", RequiredRoles: []string{"reviewer"},
	})
	if err != nil {
		t.Fatal(err)
	}

	// The AI agent holds the reviewer role but must still be refused.
	if _, err := svc.Decide(a.ID, approval.StatusApproved, "ai.reviewer.codex", "lgtm"); err == nil {
		t.Fatal("AI actor with required role decided a human-gated approval; want refusal")
	}
	// The human (also holding the role) may decide it.
	if _, err := svc.Decide(a.ID, approval.StatusApproved, "human.owner", "approved"); err != nil {
		t.Fatalf("human with required role refused: %v", err)
	}
}

func TestRequestLandingStaysHuman(t *testing.T) {
	svc, db := requestService(t)
	tk := &ticket.Ticket{Title: "feature", Status: ticket.StatusInProgress}
	if err := db.CreateTicket(tk, "tester"); err != nil {
		t.Fatal(err)
	}

	a, err := svc.Request(RequestParams{
		TicketID: tk.ID, Type: approval.TypeLandToMain, Summary: "land",
		Action: policy.Action{Type: "land_to_main", ChangeType: "documentation",
			Scope: risk.Scope{Files: []string{"docs/readme.md"}}},
	})
	if err != nil {
		t.Fatalf("Request: %v", err)
	}
	if a.Status != string(approval.StatusPending) {
		t.Errorf("landing status = %s, want pending (human-gated)", a.Status)
	}
}
