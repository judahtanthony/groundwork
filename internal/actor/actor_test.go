package actor

import "testing"

const validRegistry = `schema: groundwork_actors/v1
actors:
  - id: human.owner
    type: human
    display_name: Owner
    roles: [owner]
  - id: ai.codex.default
    type: ai_agent
    display_name: Codex Default
    runtime: codex
    capabilities:
      work_types: [technical_implementation, documentation]
`

func TestParseValid(t *testing.T) {
	reg, warnings, err := Parse([]byte(validRegistry))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if len(warnings) != 0 {
		t.Fatalf("unexpected warnings: %v", warnings)
	}
	if len(reg.Actors) != 2 {
		t.Fatalf("actors = %d, want 2", len(reg.Actors))
	}
	a, ok := reg.Get("ai.codex.default")
	if !ok || a.Runtime != "codex" || len(a.Capabilities.WorkTypes) != 2 {
		t.Fatalf("codex actor parsed wrong: %+v", a)
	}
}

func TestParseDuplicateID(t *testing.T) {
	dup := `schema: groundwork_actors/v1
actors:
  - id: x
    type: human
  - id: x
    type: human
`
	if _, _, err := Parse([]byte(dup)); err == nil {
		t.Fatal("expected duplicate id error")
	}
}

func TestParseInvalidType(t *testing.T) {
	bad := `schema: groundwork_actors/v1
actors:
  - id: x
    type: wizard
`
	if _, _, err := Parse([]byte(bad)); err == nil {
		t.Fatal("expected invalid type error")
	}
}

func TestParseSchemaMismatchWarns(t *testing.T) {
	other := `schema: groundwork_actors/v2
actors:
  - id: x
    type: human
`
	_, warnings, err := Parse([]byte(other))
	if err != nil {
		t.Fatal(err)
	}
	if len(warnings) == 0 {
		t.Fatal("expected a schema-mismatch warning")
	}
}

func TestParseEmptyRejected(t *testing.T) {
	if _, _, err := Parse([]byte("schema: groundwork_actors/v1\n")); err == nil {
		t.Fatal("expected error for empty actor list")
	}
}
