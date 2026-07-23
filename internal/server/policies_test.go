package server

import (
	"os"
	"path/filepath"
	"testing"

	"groundwork/internal/policy"
	"groundwork/internal/store/sqlite"
	"groundwork/internal/ticket"
)

const testTrustPolicy = `schema: groundwork_trust_policy/v1
auto_approve:
  - id: docs
    description: docs first
    when:
      files: ["**/*.md"]
require_human:
  - id: landing
    description: landing second
    when:
      action_types: [land_to_main]
allow_claim:
  - id: coding
    description: coding third
    when:
      actor_types: [ai_agent]
    actions: [execute]
`

const testValidationPolicy = `schema: groundwork_validation_policy/v1
templates:
  go:
    match:
      files: ["**/*.go"]
    required:
      - name: go_tests
        command: go test ./...
  docs:
    match:
      files: ["**/*.md"]
    required: []
`

func writeTestPolicies(t *testing.T, srv *Server) {
	t.Helper()
	dir := srv.proj.PoliciesDir()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	for name, body := range map[string]string{"trust.yaml": testTrustPolicy, "validation.yaml": testValidationPolicy} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

func TestPoliciesAPIListsStableRulesAndSortedValidationTemplates(t *testing.T) {
	srv, _ := newTestServer(t)
	writeTestPolicies(t, srv)
	var got policiesResponse
	if code := get(t, srv, "/api/v1/policies", &got); code != 200 {
		t.Fatalf("GET policies = %d", code)
	}
	wantIDs := []string{"landing", "docs", "coding"}
	if len(got.Rules) != len(wantIDs) {
		t.Fatalf("rules = %d, want %d", len(got.Rules), len(wantIDs))
	}
	for i, want := range wantIDs {
		if got.Rules[i].Rule.ID != want || got.Rules[i].Order != 1 {
			t.Errorf("rule[%d] = %+v, want id %q order 1", i, got.Rules[i], want)
		}
	}
	if len(got.ValidationTemplates) != 2 || got.ValidationTemplates[0].Name != "docs" || got.ValidationTemplates[1].Name != "go" {
		t.Errorf("validation templates = %+v, want docs then go", got.ValidationTemplates)
	}
}

func TestPolicyUpdateIsAppliedOnlyAfterApproval(t *testing.T) {
	srv, db := newTestServer(t)
	writeTestPolicies(t, srv)
	tk := mustCreate(t, db, &ticket.Ticket{Title: "Policy change", Status: ticket.StatusInProgress})
	set, _, err := policy.Load(srv.proj.PoliciesDir())
	if err != nil {
		t.Fatal(err)
	}
	set.Trust.RequireHuman[0].Description = "edited through the gated API"
	var queued struct {
		Approval sqlite.Approval `json:"approval"`
	}
	if code := req(t, srv, "PUT", "/api/v1/policies", policyUpdateRequest{TicketID: tk.ID, Trust: *set.Trust}, &queued); code != 202 {
		t.Fatalf("PUT policies = %d", code)
	}
	before, _, _ := policy.Load(srv.proj.PoliciesDir())
	if before.Trust.RequireHuman[0].Description != "landing second" {
		t.Fatal("policy changed before approval")
	}
	if code := req(t, srv, "POST", "/api/v1/approvals/"+queued.Approval.ID+"/approve", map[string]string{"reason": "approved policy work"}, nil); code != 200 {
		t.Fatalf("approve amendment = %d", code)
	}
	after, _, err := policy.Load(srv.proj.PoliciesDir())
	if err != nil {
		t.Fatal(err)
	}
	if after.Trust.RequireHuman[0].Description != "edited through the gated API" {
		t.Errorf("description = %q", after.Trust.RequireHuman[0].Description)
	}
}

func TestPolicySuggestionQueueAPI(t *testing.T) {
	srv, db := newTestServer(t)
	for i := 0; i < sqlite.SuggestionElevationThreshold; i++ {
		tk := mustCreate(t, db, &ticket.Ticket{Title: "clean", NodeType: ticket.NodeLeaf, Status: ticket.StatusDone, WorkType: "documentation"})
		if _, err := db.RecordValidation(sqlite.ValidationResult{TicketID: tk.ID, Name: "docs", Status: sqlite.ValidationPass}); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := db.GenerateElevationSuggestions(); err != nil {
		t.Fatal(err)
	}
	var suggestions []sqlite.PolicySuggestion
	if code := get(t, srv, "/api/v1/policies/suggestions", &suggestions); code != 200 || len(suggestions) != 1 {
		t.Fatalf("GET suggestions = %d, items=%d", code, len(suggestions))
	}
	if code := req(t, srv, "POST", "/api/v1/policies/suggestions/"+suggestions[0].ID+"/dismiss", nil, nil); code != 200 {
		t.Fatalf("dismiss suggestion = %d", code)
	}
	if code := get(t, srv, "/api/v1/policies/suggestions", &suggestions); code != 200 || len(suggestions) != 0 {
		t.Fatalf("pending after dismiss = %d, items=%d", code, len(suggestions))
	}
}
