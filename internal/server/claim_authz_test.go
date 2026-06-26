package server

import (
	"testing"

	"groundwork/internal/actor"
	"groundwork/internal/approval"
	"groundwork/internal/envelope"
	"groundwork/internal/policy"
	"groundwork/internal/risk"
	"groundwork/internal/ticket"
)

func boolp(b bool) *bool { return &b }

func TestAuthorizeEnvelopedClaim(t *testing.T) {
	srv, db := newTestServer(t)
	reg, _, _ := actor.Parse([]byte(testActorsYAML))
	// Trust allows the coding role to execute only within an envelope.
	pol := &policy.Set{Trust: &policy.TrustPolicy{AllowClaim: []policy.Rule{{
		ID: "coding-within", When: policy.Match{Roles: []string{"coding"}, WithinEnvelope: boolp(true)}, Actions: []string{"execute"},
	}}}}
	srv.SetApprovals(NewApprovalService(db, pol, reg))

	root := &ticket.Ticket{Title: "root", NodeType: ticket.NodeComposite, Status: ticket.StatusTodo, WorkType: "technical_design"}
	mustCreate(t, db, root)
	child := &ticket.Ticket{ParentID: root.ID, Title: "child", NodeType: ticket.NodeLeaf, Status: ticket.StatusInProgress, WorkType: "technical_implementation"}
	mustCreate(t, db, child)
	if err := db.UpsertEnvelope(&envelope.Envelope{
		ID: "ENV-0001", NodeID: root.ID, Status: envelope.StatusActive,
		ApprovedActions: []string{envelope.ActionExecuteChildren}, AllowedRoles: []string{"coding"},
		Planning: envelope.Planning{AllowedWorkTypes: []string{"technical_implementation"}}, RiskCeiling: "medium",
		Scope: envelope.Scope{Files: envelope.FileScope{Allow: []string{"internal/**"}}},
	}); err != nil {
		t.Fatal(err)
	}

	human := &actor.Actor{ID: "human.owner", Type: actor.TypeHuman, Roles: []string{"owner"}}
	coder := &actor.Actor{ID: "ai.coding.codex", Type: actor.TypeAIAgent, Roles: []string{"coding"}}
	planner := &actor.Actor{ID: "ai.planner.codex", Type: actor.TypeAIAgent, Roles: []string{"planner"}}
	wt := "technical_implementation"

	// Human bypasses the envelope.
	if oc, _, _ := srv.AuthorizeEnvelopedClaim(child.ID, "execute", wt, human, risk.ClassLow, []string{"anywhere.go"}); oc != ClaimAllow {
		t.Errorf("human: %s, want allow", oc)
	}
	// AI coding, in scope: trust AND envelope allow.
	if oc, _, _ := srv.AuthorizeEnvelopedClaim(child.ID, "execute", wt, coder, risk.ClassLow, []string{"internal/x.go"}); oc != ClaimAllow {
		t.Errorf("coder in-scope: %s, want allow", oc)
	}
	// AI coding, out of scope but trust-allowable: boundary crossing -> exception.
	oc, appr, err := srv.AuthorizeEnvelopedClaim(child.ID, "execute", wt, coder, risk.ClassLow, []string{"cmd/main.go"})
	if err != nil {
		t.Fatal(err)
	}
	if oc != ClaimException || appr == nil || appr.Type != string(approval.TypeException) || appr.Status != string(approval.StatusPending) {
		t.Errorf("coder out-of-scope: oc=%s appr=%v, want exception + pending exception approval", oc, appr)
	}
	// Wrong role: trust never allows -> plain deny (no exception).
	if oc, _, _ := srv.AuthorizeEnvelopedClaim(child.ID, "execute", wt, planner, risk.ClassLow, []string{"internal/x.go"}); oc != ClaimDeny {
		t.Errorf("planner: %s, want deny", oc)
	}
	// No envelope in the chain -> deny.
	orphan := &ticket.Ticket{Title: "orphan", NodeType: ticket.NodeLeaf, Status: ticket.StatusInProgress, WorkType: wt}
	mustCreate(t, db, orphan)
	if oc, _, _ := srv.AuthorizeEnvelopedClaim(orphan.ID, "execute", wt, coder, risk.ClassLow, []string{"internal/x.go"}); oc != ClaimDeny {
		t.Errorf("orphan: %s, want deny", oc)
	}
}
