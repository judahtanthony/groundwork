package policy

import (
	"testing"

	"groundwork/internal/actor"
	"groundwork/internal/risk"
)

func boolPtr(b bool) *bool { return &b }

var codex = &actor.Actor{
	ID:           "ai.codex.default",
	Type:         actor.TypeAIAgent,
	Roles:        []string{"implementer"},
	Capabilities: actor.Capabilities{WorkTypes: []string{"technical_implementation", "documentation"}},
}

func TestEvaluateReversibilityFloorForcesCritical(t *testing.T) {
	// Even with a broad auto_approve rule, an irreversible action is human-gated.
	set := &Set{Trust: &TrustPolicy{AutoApprove: []Rule{{ID: "all", When: Match{Files: []string{"**/*"}}}}}}
	d := set.Evaluate(Action{
		Type:  "execute",
		Scope: risk.Scope{Files: []string{"db/migrate.sql"}, External: true},
	})
	if d.Outcome != OutcomeRequireHuman {
		t.Fatalf("outcome = %s, want require_human", d.Outcome)
	}
	if d.RiskClass != risk.ClassCritical {
		t.Errorf("class = %s, want critical", d.RiskClass)
	}
	if d.RuleID != "" || len(d.Reasons) == 0 {
		t.Errorf("expected floor decision (no rule, with reasons), got %+v", d)
	}
}

func TestEvaluateRequireHumanBeatsAutoApprove(t *testing.T) {
	set := &Set{Trust: &TrustPolicy{
		AutoApprove:  []Rule{{ID: "docs", When: Match{Files: []string{"**/*.md"}}}},
		RequireHuman: []Rule{{ID: "land", When: Match{ActionTypes: []string{"land_to_main"}}}},
	}}
	d := set.Evaluate(Action{Type: "land_to_main", Scope: risk.Scope{Files: []string{"README.md"}}})
	if d.Outcome != OutcomeRequireHuman || d.RuleID != "land" {
		t.Fatalf("got %s/%s, want require_human/land", d.Outcome, d.RuleID)
	}
}

func TestEvaluateRequireHumanCarriesRequiredRoles(t *testing.T) {
	set := &Set{Trust: &TrustPolicy{
		RequireHuman: []Rule{{
			ID: "land-needs-staff", When: Match{ActionTypes: []string{"land_to_main"}},
			RequireRoles: []string{"staff_engineer"},
		}},
	}}
	d := set.Evaluate(Action{Type: "land_to_main", Scope: risk.Scope{Files: []string{"README.md"}}})
	if d.Outcome != OutcomeRequireHuman || d.RuleID != "land-needs-staff" {
		t.Fatalf("outcome=%v rule=%q, want require_human/land-needs-staff", d.Outcome, d.RuleID)
	}
	if len(d.RequiredRoles) != 1 || d.RequiredRoles[0] != "staff_engineer" {
		t.Errorf("required roles = %v, want [staff_engineer]", d.RequiredRoles)
	}
}

// Role is a live policy input: a role-scoped rule matches only actors that hold
// the role (ADR 0055).
func TestEvaluateRoleScopedRuleMatchesByRole(t *testing.T) {
	set := &Set{Trust: &TrustPolicy{
		RequireHuman: []Rule{{ID: "coding-gate", When: Match{Roles: []string{"coding"}}}},
	}}
	coder := &actor.Actor{ID: "ai.coding.codex", Type: actor.TypeAIAgent, Roles: []string{"coding"}}
	planner := &actor.Actor{ID: "ai.planner.codex", Type: actor.TypeAIAgent, Roles: []string{"planner"}}
	if d := set.Evaluate(Action{Type: "execute", Actor: coder, Scope: risk.Scope{Files: []string{"README.md"}}}); d.RuleID != "coding-gate" {
		t.Errorf("coding actor: rule=%q, want coding-gate", d.RuleID)
	}
	if d := set.Evaluate(Action{Type: "execute", Actor: planner, Scope: risk.Scope{Files: []string{"README.md"}}}); d.RuleID == "coding-gate" {
		t.Errorf("planner actor matched coding-only rule")
	}
}

func TestEvaluateAutoApproveDocs(t *testing.T) {
	set := &Set{Trust: &TrustPolicy{AutoApprove: []Rule{{
		ID:   "internal_docs",
		When: Match{Files: []string{"**/*.md"}, ChangeType: "documentation", MaxDiffLines: 200},
	}}}}

	// Matches: docs file, declared change type, small diff, reversible, low risk.
	ok := set.Evaluate(Action{Type: "execute", ChangeType: "documentation", DiffLines: 10,
		Scope: risk.Scope{Files: []string{"docs/x.md"}}})
	if ok.Outcome != OutcomeAutoApprove || ok.RuleID != "internal_docs" {
		t.Fatalf("got %s/%s, want auto_approve/internal_docs", ok.Outcome, ok.RuleID)
	}

	// Without a declared change_type the rule must not match (conservative);
	// with no autonomy policy it falls through to require_human.
	noType := set.Evaluate(Action{Type: "execute", DiffLines: 10,
		Scope: risk.Scope{Files: []string{"docs/x.md"}}})
	if noType.Outcome != OutcomeRequireHuman {
		t.Errorf("missing change_type outcome = %s, want require_human", noType.Outcome)
	}
}

func TestEvaluateAutonomyDefaultAndElevation(t *testing.T) {
	set := &Set{Autonomy: &AutonomyPolicy{Actions: map[string]AutonomyAction{
		"decompose": {Default: "require_human", ByWorkType: map[string]AutonomyByWorkType{
			"documentation": {Level: "auto"},
		}},
	}}}

	def := set.Evaluate(Action{Type: "decompose", WorkType: "technical_implementation",
		Scope: risk.Scope{Files: []string{"a.go"}}})
	if def.Outcome != OutcomeRequireHuman {
		t.Errorf("default outcome = %s, want require_human", def.Outcome)
	}

	elevated := set.Evaluate(Action{Type: "decompose", WorkType: "documentation",
		Scope: risk.Scope{Files: []string{"a.md"}}})
	if elevated.Outcome != OutcomeAutoApprove {
		t.Errorf("elevated outcome = %s, want auto_approve", elevated.Outcome)
	}
}

func TestEvaluateEmptySetIsConservative(t *testing.T) {
	d := (&Set{}).Evaluate(Action{Type: "execute", Scope: risk.Scope{Files: []string{"a.go"}}})
	if d.Outcome != OutcomeRequireHuman {
		t.Errorf("outcome = %s, want require_human", d.Outcome)
	}
}

func TestAuthorizeClaimAllow(t *testing.T) {
	set := &Set{Trust: &TrustPolicy{AllowClaim: []Rule{{
		ID:      "codex_medium",
		When:    Match{ActorIDs: []string{"ai.codex.default"}, WorkTypes: []string{"technical_implementation"}, RiskClassAtMost: "medium"},
		Actions: []string{"execute", "decompose"},
	}}}}
	d := set.AuthorizeClaim(Action{Type: "execute", Actor: codex, WorkType: "technical_implementation",
		Scope: risk.Scope{Files: []string{"a.go"}}})
	if d.Outcome != OutcomeAllow || d.RuleID != "codex_medium" {
		t.Fatalf("got %s/%s, want allow/codex_medium", d.Outcome, d.RuleID)
	}
}

func TestAuthorizeClaimPrefixMatch(t *testing.T) {
	// A rule keyed on the "ai" prefix authorizes the ai.codex.default instance.
	set := &Set{Trust: &TrustPolicy{AllowClaim: []Rule{{
		ID: "any_ai", When: Match{ActorIDs: []string{"ai"}}, Actions: []string{"execute"},
	}}}}
	d := set.AuthorizeClaim(Action{Type: "execute", Actor: codex, Scope: risk.Scope{Files: []string{"a.go"}}})
	if d.Outcome != OutcomeAllow {
		t.Fatalf("prefix claim outcome = %s, want allow", d.Outcome)
	}
}

func TestAuthorizeClaimDenies(t *testing.T) {
	set := &Set{Trust: &TrustPolicy{AllowClaim: []Rule{{
		ID: "codex_only", When: Match{ActorIDs: []string{"ai.codex.default"}}, Actions: []string{"execute"},
	}}}}

	// Wrong action type.
	if d := set.AuthorizeClaim(Action{Type: "land_to_main", Actor: codex, Scope: risk.Scope{Files: []string{"a.go"}}}); d.Outcome != OutcomeDeny {
		t.Errorf("wrong-action outcome = %s, want deny", d.Outcome)
	}
	// Unresolved actor.
	if d := set.AuthorizeClaim(Action{Type: "execute", Scope: risk.Scope{Files: []string{"a.go"}}}); d.Outcome != OutcomeDeny {
		t.Errorf("nil-actor outcome = %s, want deny", d.Outcome)
	}
	// No policy at all.
	if d := (&Set{}).AuthorizeClaim(Action{Type: "execute", Actor: codex}); d.Outcome != OutcomeDeny {
		t.Errorf("no-policy outcome = %s, want deny", d.Outcome)
	}
}

func TestMatchReversibleField(t *testing.T) {
	// A rule asserting reversible:false matches only irreversible scopes.
	set := &Set{Trust: &TrustPolicy{RequireHuman: []Rule{{
		ID: "irrev", When: Match{Reversible: boolPtr(false)},
	}}}}
	// Reversible action: rule doesn't apply -> falls through to (no autonomy) require_human anyway,
	// but RuleID must be empty, proving the rule did not fire.
	d := set.Evaluate(Action{Type: "execute", Scope: risk.Scope{Files: []string{"a.go"}}})
	if d.RuleID != "" {
		t.Errorf("reversible action fired rule %q, want none", d.RuleID)
	}
}

func TestGlobMatching(t *testing.T) {
	cases := []struct {
		glob, file string
		want       bool
	}{
		{"**/*.md", "README.md", true},
		{"**/*.md", "docs/a/b.md", true},
		{"**/*.md", "main.go", false},
		{"billing/**", "billing/api/charge.go", true},
		{"billing/**", "payments/x.go", false},
		{"**/.env*", "config/.env.production", true},
		{"AGENTS.md", "AGENTS.md", true},
	}
	for _, tc := range cases {
		got := anyFileMatch([]string{tc.glob}, []string{tc.file})
		if got != tc.want {
			t.Errorf("glob %q vs %q = %v, want %v", tc.glob, tc.file, got, tc.want)
		}
	}
}
