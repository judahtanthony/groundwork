package completion

import "testing"

func TestSidecarRoundTrip(t *testing.T) {
	dir := t.TempDir()
	want := &Summary{
		NodeID: "T-2001", Outcome: "Implemented the thing",
		Changed:    []string{"internal/x.go"},
		Validation: []ValidationLine{{Command: "go test ./...", Status: "pass"}},
		Decisions:  []string{"used approach A"},
		Risks:      []string{"resume needs a follow-up test"},
	}
	if err := Write(dir, want); err != nil {
		t.Fatal(err)
	}
	got, ok, err := Read(dir, "T-2001")
	if err != nil || !ok {
		t.Fatalf("read: ok=%v err=%v", ok, err)
	}
	if got.Outcome != want.Outcome || len(got.Changed) != 1 || len(got.Validation) != 1 || got.Validation[0].Status != "pass" {
		t.Errorf("round-trip mismatch: %+v", got)
	}
}

func TestReadMissing(t *testing.T) {
	if _, ok, err := Read(t.TempDir(), "T-9999"); ok || err != nil {
		t.Errorf("missing: ok=%v err=%v, want false/nil", ok, err)
	}
}
