package agentsync

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSyncPreservesExistingInstructionsAndIsIdempotent(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "AGENTS.md")
	if err := os.WriteFile(path, []byte("# Local instructions\n\nKeep this text.\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	status, err := Sync(root)
	if err != nil {
		t.Fatal(err)
	}
	if status.State != "synced" {
		t.Fatalf("state = %q, want synced", status.State)
	}
	first, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(first), "Keep this text.") || !strings.Contains(string(first), managedBlock) {
		t.Fatalf("sync did not preserve existing content:\n%s", first)
	}
	if info, err := os.Stat(path); err != nil || info.Mode().Perm() != 0o600 {
		t.Fatalf("mode = %v, err = %v; want 0600", info.Mode().Perm(), err)
	}

	if _, err := Sync(root); err != nil {
		t.Fatal(err)
	}
	second, _ := os.ReadFile(path)
	if string(second) != string(first) {
		t.Fatalf("second sync changed the file:\n%s", second)
	}
}

func TestInspectMissing(t *testing.T) {
	status, err := Inspect(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if status.State != "missing" {
		t.Fatalf("state = %q, want missing", status.State)
	}
}
