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
