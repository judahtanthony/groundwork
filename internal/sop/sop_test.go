package sop

import (
	"os"
	"path/filepath"
	"testing"
)

func TestListAndLoad(t *testing.T) {
	sops := t.TempDir()
	docsDir := filepath.Join(sops, "documentation")
	if err := os.MkdirAll(docsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(docsDir, "index.md"), []byte("# Docs SOP"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(docsDir, ".hidden"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	list, err := List(sops, "documentation")
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(list) != 1 || list[0] != filepath.Join("documentation", "index.md") {
		t.Fatalf("List = %v, want [documentation/index.md]", list)
	}

	loaded, err := Load(sops, "documentation")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if loaded[filepath.Join("documentation", "index.md")] != "# Docs SOP" {
		t.Errorf("Load content = %q", loaded)
	}
}

func TestMissingSOPDirIsEmpty(t *testing.T) {
	list, err := List(t.TempDir(), "deployment")
	if err != nil || len(list) != 0 {
		t.Fatalf("List = %v, err = %v, want empty/no-error", list, err)
	}
	if _, err := List(t.TempDir(), ""); err != nil {
		t.Errorf("empty work type should not error: %v", err)
	}
}
