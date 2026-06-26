package envelope

import (
	"testing"
)

func sample(nodeID string) *Envelope {
	return &Envelope{
		ID: "env-" + nodeID, NodeID: nodeID, Status: StatusActive,
		ApprovedBy: "human.owner", ApprovedAt: "2026-06-25T00:00:00Z",
		ApprovedActions: []string{ActionExecuteChildren, ActionLandChildToParent},
		Planning:        Planning{MaxDepth: 2, MaxChildren: 12, AllowedWorkTypes: []string{"technical_implementation"}},
		Scope:           Scope{Files: FileScope{Allow: []string{"internal/**"}, Deny: []string{".env*"}}},
		Validation:      Validation{RequiredTemplates: []string{"go"}},
		RiskCeiling:     "medium",
		AllowedRoles:    []string{"coding"},
		Escalation:      Escalation{OnUnexpectedFiles: true, OnValidationFailure: true},
	}
}

func TestSidecarRoundTrip(t *testing.T) {
	dir := t.TempDir()
	want := sample("T-2000")
	if err := Write(dir, want); err != nil {
		t.Fatal(err)
	}
	got, ok, err := Read(dir, "T-2000")
	if err != nil || !ok {
		t.Fatalf("read: ok=%v err=%v", ok, err)
	}
	if got.ID != want.ID || got.RiskCeiling != "medium" || got.Planning.MaxChildren != 12 {
		t.Errorf("round-trip mismatch: %+v", got)
	}
	if len(got.ApprovedActions) != 2 || !got.Allows(ActionExecuteChildren) {
		t.Errorf("approved actions not preserved: %v", got.ApprovedActions)
	}
}

func TestReadMissingSidecar(t *testing.T) {
	_, ok, err := Read(t.TempDir(), "T-9999")
	if err != nil || ok {
		t.Errorf("missing sidecar: ok=%v err=%v, want false/nil", ok, err)
	}
}

func TestEnvelopeMatchers(t *testing.T) {
	e := sample("T-2000")
	if !e.Allows(ActionExecuteChildren) || e.Allows(ActionDecomposeChildren) {
		t.Error("Allows mismatch")
	}
	if !e.AllowsRole("coding") || e.AllowsRole("planner") {
		t.Error("AllowsRole mismatch")
	}
	if !e.AllowsWorkType("technical_implementation") || e.AllowsWorkType("documentation") {
		t.Error("AllowsWorkType mismatch")
	}
}
