package policy

import (
	"os"
	"path/filepath"
	"testing"
)

const trustYAML = `schema: groundwork_trust_policy/v1
auto_approve:
  - id: internal_docs
    when:
      files: ["**/*.md"]
      change_type: documentation
      max_diff_lines: 200
allow_claim:
  - id: default_codex_medium_risk
    when:
      actor_ids: [ai.codex.default]
      work_types: [technical_implementation, documentation]
      risk_class_at_most: medium
    actions: [execute, decompose]
require_human:
  - id: landing_to_main_v1
    when:
      action_types: [land_to_main]
`

const autonomyYAML = `schema: groundwork_autonomy_policy/v1
actions:
  execute:
    default: require_human
  decompose:
    default: require_human
    by_work_type:
      documentation:
        level: auto
`

const validationYAML = `schema: groundwork_validation_policy/v1
templates:
  go:
    match:
      files: ["**/*.go"]
    required:
      - name: go_tests
        command: "go test ./..."
    landing_risk_floor: low
`

func TestParseTrustValid(t *testing.T) {
	p, warnings, err := ParseTrust([]byte(trustYAML))
	if err != nil {
		t.Fatalf("ParseTrust: %v", err)
	}
	if len(warnings) != 0 {
		t.Errorf("unexpected warnings: %v", warnings)
	}
	if len(p.AutoApprove) != 1 || len(p.AllowClaim) != 1 || len(p.RequireHuman) != 1 {
		t.Fatalf("unexpected rule counts: %+v", p)
	}
	if p.AllowClaim[0].When.RiskClassAtMost != "medium" {
		t.Errorf("risk_class_at_most not parsed: %+v", p.AllowClaim[0])
	}
}

func TestParseTrustDuplicateID(t *testing.T) {
	dup := `schema: groundwork_trust_policy/v1
auto_approve:
  - id: x
    when: {files: ["*"]}
require_human:
  - id: x
    when: {action_types: [land_to_main]}
`
	if _, _, err := ParseTrust([]byte(dup)); err == nil {
		t.Fatal("expected duplicate id error")
	}
}

func TestParseTrustInvalidRiskClass(t *testing.T) {
	bad := `schema: groundwork_trust_policy/v1
allow_claim:
  - id: x
    when: {risk_class_at_most: extreme}
`
	if _, _, err := ParseTrust([]byte(bad)); err == nil {
		t.Fatal("expected invalid risk class error")
	}
}

func TestParseAutonomyInvalidLevel(t *testing.T) {
	bad := `schema: groundwork_autonomy_policy/v1
actions:
  execute:
    default: sometimes
`
	if _, _, err := ParseAutonomy([]byte(bad)); err == nil {
		t.Fatal("expected invalid autonomy level error")
	}
}

func TestParseValidationInvalidFloor(t *testing.T) {
	bad := `schema: groundwork_validation_policy/v1
templates:
  go:
    landing_risk_floor: nope
`
	if _, _, err := ParseValidation([]byte(bad)); err == nil {
		t.Fatal("expected invalid landing_risk_floor error")
	}
}

func TestSchemaMismatchWarns(t *testing.T) {
	_, warnings, err := ParseAutonomy([]byte("schema: wrong/v1\nactions: {}\n"))
	if err != nil {
		t.Fatalf("ParseAutonomy: %v", err)
	}
	if len(warnings) == 0 {
		t.Error("expected a schema mismatch warning")
	}
}

func TestLoadDispatchesBySchema(t *testing.T) {
	dir := t.TempDir()
	write := func(name, body string) {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("trust.yaml", trustYAML)
	write("autonomy.yaml", autonomyYAML)
	write("validation.yaml", validationYAML)

	set, warnings, err := Load(dir)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(warnings) != 0 {
		t.Errorf("unexpected warnings: %v", warnings)
	}
	if set.Trust == nil || set.Autonomy == nil || set.Validation == nil {
		t.Fatalf("Load did not populate all policies: %+v", set)
	}
}

func TestLoadMissingDir(t *testing.T) {
	set, warnings, err := Load(filepath.Join(t.TempDir(), "absent"))
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if set.Trust != nil || len(warnings) == 0 {
		t.Errorf("expected empty set + warning, got %+v / %v", set, warnings)
	}
}

func TestLoadDuplicateSchemaErrors(t *testing.T) {
	dir := t.TempDir()
	_ = os.WriteFile(filepath.Join(dir, "trust.yaml"), []byte(trustYAML), 0o644)
	_ = os.WriteFile(filepath.Join(dir, "trust2.yaml"), []byte(trustYAML), 0o644)
	if _, _, err := Load(dir); err == nil {
		t.Fatal("expected duplicate-schema error")
	}
}
