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
