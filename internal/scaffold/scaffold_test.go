package scaffold

import (
	"os"
	"path/filepath"
	"testing"

	"groundwork/internal/actor"
	"groundwork/internal/config"
	"groundwork/internal/policy"
)

func TestInitCreatesTreeWithoutDB(t *testing.T) {
	root := t.TempDir()

	res, err := Init(root)
	if err != nil {
		t.Fatalf("Init: %v", err)
	}
	if res.AlreadyInitialized {
		t.Fatal("fresh root reported AlreadyInitialized")
	}

	want := []string{
		".groundwork/config.yaml",
		".groundwork/actors.yaml",
		".groundwork/WORKFLOW.md",
		".groundwork/policies/trust.yaml",
		".groundwork/policies/validation.yaml",
		".groundwork/policies/autonomy.yaml",
		".groundwork/.gitignore",
	}
	for _, rel := range want {
		if _, err := os.Stat(filepath.Join(root, rel)); err != nil {
			t.Errorf("expected %s: %v", rel, err)
		}
	}

	// The SQLite database must NOT be created by init.
	if _, err := os.Stat(filepath.Join(root, config.GroundworkDir, "state.sqlite")); !os.IsNotExist(err) {
		t.Errorf("init created state.sqlite (err=%v); it must be lazy", err)
	}

	// The written config must parse cleanly.
	data, err := os.ReadFile(filepath.Join(root, ".groundwork", "config.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if _, warnings, err := config.Parse(data); err != nil || len(warnings) != 0 {
		t.Fatalf("scaffolded config did not parse cleanly: warnings=%v err=%v", warnings, err)
	}

	// The scaffolded actor registry must be valid (catches template typos).
	reg, warnings, err := actor.Load(filepath.Join(root, ".groundwork", "actors.yaml"))
	if err != nil || len(warnings) != 0 {
		t.Fatalf("scaffolded actors.yaml invalid: warnings=%v err=%v", warnings, err)
	}
	if _, ok := reg.Get("human.owner"); !ok {
		t.Error("scaffolded registry missing human.owner")
	}
	if _, ok := reg.Get("ai.codex.default"); !ok {
		t.Error("scaffolded registry missing ai.codex.default")
	}

	// The scaffolded policies must load cleanly (catches template typos and
	// schema drift).
	set, pwarn, err := policy.Load(filepath.Join(root, ".groundwork", "policies"))
	if err != nil || len(pwarn) != 0 {
		t.Fatalf("scaffolded policies invalid: warnings=%v err=%v", pwarn, err)
	}
	if set.Trust == nil || set.Autonomy == nil || set.Validation == nil {
		t.Errorf("scaffolded policies incomplete: %+v", set)
	}
}

func TestInitIsIdempotent(t *testing.T) {
	root := t.TempDir()
	if _, err := Init(root); err != nil {
		t.Fatalf("first Init: %v", err)
	}

	// Tamper with a file to prove the second run does not clobber it.
	cfgPath := filepath.Join(root, ".groundwork", "config.yaml")
	if err := os.WriteFile(cfgPath, []byte("schema: groundwork_config/v1\nmax_concurrency: 9\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	res, err := Init(root)
	if err != nil {
		t.Fatalf("second Init: %v", err)
	}
	if !res.AlreadyInitialized {
		t.Fatal("second Init did not report AlreadyInitialized")
	}
	data, _ := os.ReadFile(cfgPath)
	if !contains(string(data), "max_concurrency: 9") {
		t.Fatalf("second Init clobbered existing config: %s", data)
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || indexOf(s, sub) >= 0)
}

func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
