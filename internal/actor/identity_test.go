package actor

import "testing"

func TestMatchID(t *testing.T) {
	cases := []struct {
		pattern, id string
		want        bool
	}{
		{"ai", "ai", true},
		{"ai", "ai.codex.default", true},
		{"ai.codex", "ai.codex.default", true},
		{"ai", "ainsley", false},       // not a tier boundary
		{"ai.codex", "ai.docs", false}, // diverging tier
		{"human", "human.owner", true},
		{"", "ai.codex", false},                 // empty pattern matches nothing
		{"ai.codex.default", "ai.codex", false}, // pattern deeper than id
	}
	for _, tc := range cases {
		if got := MatchID(tc.pattern, tc.id); got != tc.want {
			t.Errorf("MatchID(%q,%q) = %v, want %v", tc.pattern, tc.id, got, tc.want)
		}
	}
}

func testRegistry() *Registry {
	return &Registry{
		Schema: SchemaVersion,
		Actors: []Actor{
			{ID: "human.owner", Type: TypeHuman, Roles: []string{"owner"},
				Capabilities: Capabilities{WorkTypes: []string{"*"}, Approve: []string{"*"}, Review: []string{"*"}}},
			{ID: "ai.codex.default", Type: TypeAIAgent, Roles: []string{"implementer"},
				Capabilities: Capabilities{WorkTypes: []string{"technical_implementation", "documentation"}, Review: []string{"documentation"}}},
		},
	}
}

func TestResolveExactAndPrefix(t *testing.T) {
	r := testRegistry()

	if a, ok := r.Resolve("ai.codex.default"); !ok || a.ID != "ai.codex.default" {
		t.Errorf("exact resolve = %v,%v", a, ok)
	}
	if a, ok := r.Resolve("ai"); !ok || a.ID != "ai.codex.default" {
		t.Errorf("prefix resolve = %v,%v, want ai.codex.default", a, ok)
	}
	if _, ok := r.Resolve("nobody"); ok {
		t.Error("resolve of unknown should fail")
	}
	if _, ok := r.Resolve(""); ok {
		t.Error("resolve of empty should fail")
	}
}

func TestMatch(t *testing.T) {
	r := testRegistry()
	if got := r.Match("ai"); len(got) != 1 || got[0].ID != "ai.codex.default" {
		t.Errorf("Match(ai) = %v, want [ai.codex.default]", got)
	}
	if got := r.Match("human"); len(got) != 1 {
		t.Errorf("Match(human) = %v, want one", got)
	}
}

func TestCapabilityPredicates(t *testing.T) {
	r := testRegistry()
	owner, _ := r.Get("human.owner")
	codex, _ := r.Get("ai.codex.default")

	if !owner.CanClaim("anything") { // "*" wildcard
		t.Error("owner with * work_types should claim anything")
	}
	if !codex.CanClaim("documentation") || codex.CanClaim("deployment") {
		t.Error("codex claim scope wrong")
	}
	if !codex.CanReview("documentation") || codex.CanApprove("documentation") {
		t.Error("codex should review (not approve) documentation")
	}
	if !owner.HasRole("owner") || codex.HasRole("owner") {
		t.Error("role check wrong")
	}
}
