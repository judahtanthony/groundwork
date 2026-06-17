package config

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestParseDefaults(t *testing.T) {
	cfg, warnings, err := Parse([]byte("schema: " + SchemaVersion + "\n"))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if len(warnings) != 0 {
		t.Fatalf("unexpected warnings: %v", warnings)
	}
	if cfg.Runtime != "codex" {
		t.Errorf("runtime default = %q, want codex", cfg.Runtime)
	}
	if cfg.MaxConcurrency != 4 {
		t.Errorf("max_concurrency default = %d, want 4", cfg.MaxConcurrency)
	}
	if cfg.Lease.TTL.Duration() != 90*time.Second {
		t.Errorf("lease.ttl default = %v, want 90s", cfg.Lease.TTL.Duration())
	}
}

func TestParseOverridesAndDurations(t *testing.T) {
	data := []byte(`schema: groundwork_config/v1
max_concurrency: 8
lease:
  ttl: 2m
  heartbeat: 45s
`)
	cfg, warnings, err := Parse(data)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if len(warnings) != 0 {
		t.Fatalf("unexpected warnings: %v", warnings)
	}
	if cfg.MaxConcurrency != 8 {
		t.Errorf("max_concurrency = %d, want 8", cfg.MaxConcurrency)
	}
	if cfg.Lease.TTL.Duration() != 2*time.Minute {
		t.Errorf("lease.ttl = %v, want 2m", cfg.Lease.TTL.Duration())
	}
	if cfg.Lease.Heartbeat.Duration() != 45*time.Second {
		t.Errorf("lease.heartbeat = %v, want 45s", cfg.Lease.Heartbeat.Duration())
	}
}

func TestParseUnknownKeyWarns(t *testing.T) {
	cfg, warnings, err := Parse([]byte("schema: groundwork_config/v1\nbogus: 1\n"))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if cfg == nil {
		t.Fatal("expected config despite unknown key")
	}
	if len(warnings) != 1 {
		t.Fatalf("want 1 warning, got %v", warnings)
	}
}

func TestDiscoverWalksUpward(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, GroundworkDir), 0o755); err != nil {
		t.Fatal(err)
	}
	nested := filepath.Join(root, "a", "b", "c")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatal(err)
	}

	got, err := Discover(nested, "")
	if err != nil {
		t.Fatalf("Discover: %v", err)
	}
	// macOS temp dirs may be symlinked; compare resolved paths.
	wantResolved, _ := filepath.EvalSymlinks(root)
	gotResolved, _ := filepath.EvalSymlinks(got)
	if gotResolved != wantResolved {
		t.Fatalf("Discover = %q, want %q", gotResolved, wantResolved)
	}
}

func TestDiscoverNotFound(t *testing.T) {
	_, err := Discover(t.TempDir(), "")
	if !errors.Is(err, ErrProjectNotFound) {
		t.Fatalf("want ErrProjectNotFound, got %v", err)
	}
}

func TestOpenLoadsConfig(t *testing.T) {
	root := t.TempDir()
	gw := filepath.Join(root, GroundworkDir)
	if err := os.MkdirAll(gw, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(gw, "config.yaml"), []byte("schema: "+SchemaVersion+"\nmax_concurrency: 2\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	p, err := Open(root, "")
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	if p.Config.MaxConcurrency != 2 {
		t.Errorf("max_concurrency = %d, want 2", p.Config.MaxConcurrency)
	}
	if p.DBPath() != filepath.Join(gw, "state.sqlite") {
		t.Errorf("DBPath = %q", p.DBPath())
	}
}
