package runtime

import (
	"context"
	"testing"
)

func TestSelectChoosesRuntime(t *testing.T) {
	cases := map[string]string{"": "stub", "stub": "stub", "codex": "codex"}
	for name, want := range cases {
		rt, err := Select(name, Config{})
		if err != nil {
			t.Fatalf("Select(%q): %v", name, err)
		}
		if rt.Name() != want {
			t.Errorf("Select(%q).Name() = %q, want %q", name, rt.Name(), want)
		}
	}
	if _, err := Select("bogus", Config{}); err == nil {
		t.Error("Select(bogus): expected error")
	}
}

func TestNewCodexDefaultsCommand(t *testing.T) {
	c := NewCodex(Config{})
	if c.cfg.Command != "codex" {
		t.Errorf("default command = %q, want codex", c.cfg.Command)
	}
	if c.Name() != "codex" {
		t.Errorf("name = %q", c.Name())
	}
}

// TestCodexRunDelegatesToLauncherWithActorConfig proves the adapter resolves the
// effective model from the coordinator's Spec and hands the actor-configured
// attempt to the launcher (T-0501 acceptance: launch accepts actor config).
func TestCodexRunDelegatesToLauncherWithActorConfig(t *testing.T) {
	var gotSpec Spec
	var gotCfg Config
	c := NewCodex(Config{Model: "fallback-model", Sandbox: "workspace-write"}).
		WithLauncher(func(ctx context.Context, spec Spec, sink Sink, cfg Config) (Result, error) {
			gotSpec, gotCfg = spec, cfg
			sink(Event{Type: "claimed"})
			return Result{Status: "produced"}, nil
		})

	spec := Spec{RunID: "R-1", TicketID: "T-1", ActorID: "ai.codex.default", Model: "actor-model", Workspace: "/wt"}
	var events int
	res, err := c.Run(context.Background(), spec, func(Event) { events++ })
	if err != nil {
		t.Fatal(err)
	}
	if res.Status != "produced" {
		t.Errorf("status = %q", res.Status)
	}
	if gotSpec.ActorID != "ai.codex.default" || gotSpec.Workspace != "/wt" {
		t.Errorf("launcher got wrong actor config: %+v", gotSpec)
	}
	if gotSpec.Model != "actor-model" {
		t.Errorf("model = %q, want the Spec's actor-selected model", gotSpec.Model)
	}
	if gotCfg.Sandbox != "workspace-write" {
		t.Errorf("sandbox = %q", gotCfg.Sandbox)
	}
	if events == 0 {
		t.Error("expected events emitted to sink")
	}
}

// TestCodexRunFallsBackToConfigModel checks an empty Spec model resolves to the
// adapter's configured default.
func TestCodexRunFallsBackToConfigModel(t *testing.T) {
	var gotModel string
	c := NewCodex(Config{Model: "fallback-model"}).
		WithLauncher(func(ctx context.Context, spec Spec, sink Sink, cfg Config) (Result, error) {
			gotModel = spec.Model
			return Result{}, nil
		})
	if _, err := c.Run(context.Background(), Spec{TicketID: "T-1"}, nil); err != nil {
		t.Fatal(err)
	}
	if gotModel != "fallback-model" {
		t.Errorf("model = %q, want fallback-model", gotModel)
	}
}

// TestCodexShellLaunchRunsRecordsOnly verifies the default (pre-T-0502) launcher
// keeps the coordinator loop functional without spawning a process.
func TestCodexShellLaunchRunsRecordsOnly(t *testing.T) {
	c := NewCodex(Config{})
	var types []string
	res, err := c.Run(context.Background(), Spec{TicketID: "T-1"}, func(e Event) { types = append(types, e.Type) })
	if err != nil {
		t.Fatal(err)
	}
	if res.Status != "produced" {
		t.Errorf("status = %q, want produced", res.Status)
	}
	if len(types) == 0 || types[0] != "claimed" {
		t.Errorf("event sequence = %v", types)
	}
}
