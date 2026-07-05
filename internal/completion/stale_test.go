package completion

import "testing"

func TestStaleDetection(t *testing.T) {
	sum := &Summary{NodeID: "T-1", Outcome: "produced", Changed: []string{"a.go", "b.go"}}

	// Current: same changed set, not reworked.
	if stale, _ := Stale(sum, []string{"b.go", "a.go"}, "review"); stale {
		t.Error("matching set in review should not be stale")
	}
	// Reworked after the summary.
	if stale, reason := Stale(sum, []string{"a.go", "b.go"}, "rework"); !stale || reason == "" {
		t.Errorf("rework should be stale: stale=%v reason=%q", stale, reason)
	}
	// Changed set diverged.
	if stale, reason := Stale(sum, []string{"a.go", "c.go"}, "review"); !stale || reason == "" {
		t.Errorf("diverged set should be stale: stale=%v reason=%q", stale, reason)
	}
	// A nil summary is never stale.
	if stale, _ := Stale(nil, []string{"a.go"}, "rework"); stale {
		t.Error("nil summary should not be stale")
	}
}
