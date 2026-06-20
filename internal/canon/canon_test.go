package canon

import (
	"path/filepath"
	"testing"
)

func TestJournalAppendReadAndRatify(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "journal")

	if err := Append(dir, "T-0001", "considered option A"); err != nil {
		t.Fatalf("Append: %v", err)
	}
	if err := Ratify(dir, "T-0001", "decompose", "plan accepted"); err != nil {
		t.Fatalf("Ratify: %v", err)
	}
	entries, err := Read(dir, "T-0001")
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("entries = %d, want 2", len(entries))
	}
	if entries[1].Gate != "decompose" {
		t.Errorf("ratify gate = %q, want decompose", entries[1].Gate)
	}
}

func TestReadMissingJournalIsEmpty(t *testing.T) {
	entries, err := Read(t.TempDir(), "nope")
	if err != nil || len(entries) != 0 {
		t.Fatalf("entries=%v err=%v, want empty/no-error", entries, err)
	}
}

func TestReconcileDeduplicates(t *testing.T) {
	got := Reconcile("shared line\nparent line", []string{"shared line\nchild a", "child b\nshared line"})
	want := "shared line\nparent line\nchild a\nchild b"
	if got != want {
		t.Errorf("Reconcile =\n%q\nwant\n%q", got, want)
	}
}
