package runtime

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// writeScript creates an executable shell script that writes a file and prints
// output, standing in for the codex agent CLI.
func writeScript(t *testing.T, dir, body string) string {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("shell-script stub is POSIX-only")
	}
	path := filepath.Join(dir, "fake-codex.sh")
	if err := os.WriteFile(path, []byte("#!/bin/sh\n"+body), 0o755); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestExecLauncherRunsInWorkspaceAndStreams(t *testing.T) {
	scriptDir := t.TempDir()
	ws := t.TempDir()
	script := writeScript(t, scriptDir, "echo starting; printf 'package x\\n' > feature.go; echo finished\n")

	c := NewCodex(Config{Command: script}).WithExec()
	var types, msgs []string
	res, err := c.Run(context.Background(), Spec{RunID: "R-1", TicketID: "T-1", Workspace: ws},
		func(e Event) { types = append(types, e.Type); msgs = append(msgs, e.Message) })
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if res.Status != "produced" {
		t.Fatalf("status = %q, want produced", res.Status)
	}
	// The agent ran with cwd = the workspace (the file landed there).
	if _, err := os.Stat(filepath.Join(ws, "feature.go")); err != nil {
		t.Fatalf("agent did not run in the workspace: %v", err)
	}
	// Output was streamed as events.
	if !contains(types, "claimed") || !contains(types, "working") || !contains(types, "output") || !contains(types, "produced") {
		t.Fatalf("event types = %v", types)
	}
	if !contains(msgs, "starting") || !contains(msgs, "finished") {
		t.Fatalf("output not streamed: %v", msgs)
	}
}

func TestCodexArgsBuildsNonInteractiveExec(t *testing.T) {
	got := codexArgs(Config{Sandbox: "workspace-write", Args: []string{"--skip-git-repo-check"}},
		Spec{Model: "gpt-x", Prompt: "do the thing"})
	want := []string{"exec", "--model", "gpt-x", "--sandbox", "workspace-write", "--skip-git-repo-check", "do the thing"}
	if len(got) != len(want) {
		t.Fatalf("args = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("args = %v, want %v", got, want)
		}
	}
	// The prompt is always the final argv element (a single string; newlines are safe).
	if got[len(got)-1] != "do the thing" {
		t.Errorf("prompt not last: %v", got)
	}
	// exec is present even with no model/prompt (headless subcommand).
	if bare := codexArgs(Config{}, Spec{}); len(bare) == 0 || bare[0] != "exec" {
		t.Errorf("bare args = %v, want [exec]", bare)
	}
}

func TestExecLauncherFailingCommandReturnsError(t *testing.T) {
	scriptDir := t.TempDir()
	ws := t.TempDir()
	script := writeScript(t, scriptDir, "echo oops 1>&2; exit 3\n")
	c := NewCodex(Config{Command: script}).WithExec()
	res, err := c.Run(context.Background(), Spec{RunID: "R-2", TicketID: "T-2", Workspace: ws}, nil)
	if err == nil {
		t.Fatal("expected error for non-zero exit")
	}
	if res.Status != "failed" {
		t.Errorf("status = %q, want failed", res.Status)
	}
}

func TestValidateWorkspaceContainment(t *testing.T) {
	root := t.TempDir()
	inside := filepath.Join(root, "R-1")
	if err := os.MkdirAll(inside, 0o755); err != nil {
		t.Fatal(err)
	}
	if _, err := validateWorkspace(inside, root); err != nil {
		t.Errorf("contained workspace rejected: %v", err)
	}
	// Empty workspace is rejected.
	if _, err := validateWorkspace("", root); err == nil {
		t.Error("expected error for empty workspace")
	}
	// A workspace outside the worktree root is rejected (no agent in the repo root).
	outside := t.TempDir()
	if _, err := validateWorkspace(outside, root); err == nil {
		t.Error("expected containment error for workspace outside the root")
	}
}

// fakeProvider records the provisioning call and returns a prepared directory.
type fakeProvider struct {
	dir          string
	gotRun       string
	gotBase      string // base passed to Provision
	gotDiffBase  string // base passed to Diff
	checkpointed bool
}

func (f *fakeProvider) Provision(runID, base string) (string, error) {
	f.gotRun, f.gotBase = runID, base
	return f.dir, nil
}

func (f *fakeProvider) Diff(runID, base string) ([]string, string, error) {
	f.gotDiffBase = base
	return []string{"feature.go"}, "diff --git a/feature.go b/feature.go\n", nil
}

func (f *fakeProvider) Checkpoint(runID, message string) (string, error) {
	f.checkpointed = true
	return "abc123", nil
}

func TestCodexProvisionsWorkspaceBeforeLaunch(t *testing.T) {
	ws := t.TempDir()
	fp := &fakeProvider{dir: ws}
	var launchedIn string
	c := NewCodex(Config{}).
		WithWorkspace(fp, func(spec Spec) string { return "deadbeef" }).
		WithLauncher(func(ctx context.Context, spec Spec, sink Sink, cfg Config) (Result, error) {
			launchedIn = spec.Workspace
			return Result{Status: "produced"}, nil
		})

	res, err := c.Run(context.Background(), Spec{RunID: "R-9", TicketID: "T-9"}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if fp.gotRun != "R-9" || fp.gotBase != "deadbeef" {
		t.Errorf("provider called with run=%q base=%q", fp.gotRun, fp.gotBase)
	}
	if launchedIn != ws {
		t.Errorf("launcher workspace = %q, want provisioned %q", launchedIn, ws)
	}
	// The run's changed-file set is captured from the worktree after launch.
	if len(res.ChangedFiles) != 1 || res.ChangedFiles[0] != "feature.go" {
		t.Errorf("ChangedFiles = %v, want [feature.go]", res.ChangedFiles)
	}
	// The work is checkpointed on the run branch for landing/resume.
	if !fp.checkpointed {
		t.Error("expected a checkpoint commit after the run")
	}
}

// TestCodexDiffsAgainstIntegrationBaseNotProvisionBase covers review finding #3: a
// resumed run is provisioned from a prior checkpoint but its diff must be measured
// against the integration base, so files an interrupted run changed are still
// captured (and seen by envelope enforcement).
func TestCodexDiffsAgainstIntegrationBaseNotProvisionBase(t *testing.T) {
	ws := t.TempDir()
	fp := &fakeProvider{dir: ws}
	c := NewCodex(Config{}).
		WithWorkspace(fp, func(spec Spec) string { return "checkpoint-ref" }). // resume base
		WithDiffBase(func(spec Spec) string { return "integration-base" }).
		WithLauncher(func(ctx context.Context, spec Spec, sink Sink, cfg Config) (Result, error) {
			return Result{Status: "produced"}, nil
		})
	if _, err := c.Run(context.Background(), Spec{RunID: "R-9", TicketID: "T-9"}, nil); err != nil {
		t.Fatal(err)
	}
	if fp.gotBase != "checkpoint-ref" {
		t.Errorf("provision base = %q, want the resume checkpoint", fp.gotBase)
	}
	if fp.gotDiffBase != "integration-base" {
		t.Fatalf("diff base = %q, want the integration base (not the checkpoint)", fp.gotDiffBase)
	}
}

func contains(s []string, want string) bool {
	for _, v := range s {
		if v == want {
			return true
		}
	}
	return false
}
