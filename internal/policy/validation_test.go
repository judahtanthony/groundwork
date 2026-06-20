package policy

import "testing"

func validationPolicy() *ValidationPolicy {
	return &ValidationPolicy{
		Templates: map[string]ValidationTemplate{
			"go": {
				Match:    ValidationMatch{Files: []string{"**/*.go"}},
				Required: []ValidationCheck{{Name: "go_tests", Command: "go test ./..."}},
			},
			"docs": {
				Match:            ValidationMatch{Files: []string{"**/*.md"}},
				Required:         nil,
				LandingRiskFloor: "low",
			},
		},
	}
}

func TestRequiredChecksMatchesByFile(t *testing.T) {
	vp := validationPolicy()
	got := vp.RequiredCheckNames([]string{"internal/x.go"})
	if len(got) != 1 || got[0] != "go_tests" {
		t.Fatalf("go file required = %v, want [go_tests]", got)
	}
	if names := vp.RequiredCheckNames([]string{"README.md"}); len(names) != 0 {
		t.Errorf("docs file required = %v, want none", names)
	}
	// Mixed change requires the go check.
	if names := vp.RequiredCheckNames([]string{"README.md", "main.go"}); len(names) != 1 {
		t.Errorf("mixed required = %v, want [go_tests]", names)
	}
}

func TestLandingRiskFloor(t *testing.T) {
	vp := validationPolicy()
	if floor := vp.LandingRiskFloor([]string{"README.md"}); floor != "low" {
		t.Errorf("docs floor = %q, want low", floor)
	}
	if floor := vp.LandingRiskFloor([]string{"main.go"}); floor != "" {
		t.Errorf("go floor = %q, want empty", floor)
	}
}

func TestNilValidationPolicy(t *testing.T) {
	var vp *ValidationPolicy
	if got := vp.RequiredCheckNames([]string{"x.go"}); len(got) != 0 {
		t.Errorf("nil policy required = %v, want none", got)
	}
}
