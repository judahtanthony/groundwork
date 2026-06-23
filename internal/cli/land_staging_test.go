package cli

import (
	"strings"
	"testing"
)

type fakeStager struct {
	staged      bool
	uncommitted bool
	addAllCalls int
}

func (f *fakeStager) HasStagedChanges() (bool, error) { return f.staged, nil }
func (f *fakeStager) HasUncommitted() (bool, error)   { return f.uncommitted, nil }
func (f *fakeStager) AddAll() error                   { f.addAllCalls++; return nil }

type fakePreview struct {
	staged bool
	diff   string
}

func (f *fakePreview) HasStagedChanges() (bool, error) { return f.staged, nil }
func (f *fakePreview) StagedDiff() (string, error)     { return f.diff, nil }

func TestPreviewLandingShowsStagedDiff(t *testing.T) {
	ctx, out, _ := newTestCtx()
	repo := &fakePreview{staged: true, diff: "diff --git a/x b/x\n+added line\n"}
	if err := previewLanding(ctx, "T-0001", repo); err != nil {
		t.Fatalf("previewLanding: %v", err)
	}
	got := out.String()
	if !strings.Contains(got, "T-0001") || !strings.Contains(got, "+added line") {
		t.Errorf("preview missing id or diff:\n%s", got)
	}
}

func TestPreviewLandingNothingStaged(t *testing.T) {
	ctx, out, _ := newTestCtx()
	repo := &fakePreview{staged: false}
	if err := previewLanding(ctx, "T-0001", repo); err != nil {
		t.Fatalf("previewLanding: %v", err)
	}
	if !strings.Contains(out.String(), "No staged changes") {
		t.Errorf("expected no-staged-changes hint, got:\n%s", out.String())
	}
}

func TestResolveLandStaging(t *testing.T) {
	tests := []struct {
		name        string
		all         bool
		staged      bool
		uncommitted bool
		input       string
		wantAddAll  int
	}{
		{name: "all flag stages everything", all: true, wantAddAll: 1},
		{name: "already staged: leave the index", staged: true, uncommitted: true, wantAddAll: 0},
		{name: "clean tree: nothing to stage", wantAddAll: 0},
		{name: "empty index, prompt yes", uncommitted: true, input: "y\n", wantAddAll: 1},
		{name: "empty index, prompt default (enter)", uncommitted: true, input: "\n", wantAddAll: 1},
		{name: "empty index, EOF/non-interactive defaults no", uncommitted: true, input: "", wantAddAll: 0},
		{name: "empty index, prompt no", uncommitted: true, input: "n\n", wantAddAll: 0},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			f := &fakeStager{staged: tc.staged, uncommitted: tc.uncommitted}
			var out strings.Builder
			if err := resolveLandStaging(f, tc.all, strings.NewReader(tc.input), &out); err != nil {
				t.Fatalf("resolveLandStaging: %v", err)
			}
			if f.addAllCalls != tc.wantAddAll {
				t.Errorf("AddAll calls = %d, want %d", f.addAllCalls, tc.wantAddAll)
			}
		})
	}
}
