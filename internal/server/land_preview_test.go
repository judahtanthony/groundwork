package server

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"groundwork/internal/ticket"
)

// With a staged change, the preview reports staged=true and returns the staged
// diff — the same view as gw ticket land --preview.
func TestLandPreviewReturnsStagedDiff(t *testing.T) {
	srv, db, root := newGitServer(t)
	tk := &ticket.Ticket{Title: "preview me", Status: ticket.StatusReview, WorkType: "documentation"}
	if err := db.CreateTicket(tk, "tester"); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "NOTES.md"), []byte("hello preview\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	runGit(t, root, "add", "NOTES.md")

	var out struct {
		ID     string `json:"id"`
		Staged bool   `json:"staged"`
		Diff   string `json:"diff"`
	}
	if code := get(t, srv, "/api/v1/tickets/"+tk.ID+"/land/preview", &out); code != 200 {
		t.Fatalf("preview = %d, want 200", code)
	}
	if !out.Staged {
		t.Errorf("staged = false, want true")
	}
	if out.ID != tk.ID {
		t.Errorf("id = %q, want %q", out.ID, tk.ID)
	}
	for _, want := range []string{"NOTES.md", "hello preview"} {
		if !strings.Contains(out.Diff, want) {
			t.Errorf("diff missing %q:\n%s", want, out.Diff)
		}
	}
}

// With nothing staged, the preview reports staged=false and an empty diff.
func TestLandPreviewNothingStaged(t *testing.T) {
	srv, db, _ := newGitServer(t)
	tk := &ticket.Ticket{Title: "nothing staged", Status: ticket.StatusReview, WorkType: "documentation"}
	if err := db.CreateTicket(tk, "tester"); err != nil {
		t.Fatal(err)
	}
	var out struct {
		Staged bool   `json:"staged"`
		Diff   string `json:"diff"`
	}
	if code := get(t, srv, "/api/v1/tickets/"+tk.ID+"/land/preview", &out); code != 200 {
		t.Fatalf("preview = %d, want 200", code)
	}
	if out.Staged || out.Diff != "" {
		t.Errorf("staged=%v diff=%q, want false/empty", out.Staged, out.Diff)
	}
}

// A land_to_main approval renders the staged diff inline in the inbox so the
// human can inspect the landing before deciding (T-1065).
func TestApprovalsInboxInlineLandPreview(t *testing.T) {
	srv, db, root := newGitServer(t)
	pendingApproval(t, db) // a pending land_to_main approval
	if err := os.WriteFile(filepath.Join(root, "CHANGES.md"), []byte("diff me please\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	runGit(t, root, "add", "CHANGES.md")

	body := getHTML(t, srv, "/approvals").Body.String()
	for _, want := range []string{"Preview landing diff", `class="gw-diff"`, "CHANGES.md", "diff me please"} {
		if !strings.Contains(body, want) {
			t.Errorf("inbox land preview missing %q", want)
		}
	}
}

// An unknown ticket id is a 404 through the store-error mapping.
func TestLandPreviewUnknownTicket(t *testing.T) {
	srv, _, _ := newGitServer(t)
	if code := get(t, srv, "/api/v1/tickets/T-9999/land/preview", nil); code != 404 {
		t.Fatalf("preview unknown = %d, want 404", code)
	}
}

// Outside a git work tree the preview is a 400 not_a_repo, matching the CLI.
func TestLandPreviewNotARepo(t *testing.T) {
	srv, db := newTestServer(t) // temp dir, no git
	tk := &ticket.Ticket{Title: "no repo", Status: ticket.StatusReview, WorkType: "documentation"}
	if err := db.CreateTicket(tk, "tester"); err != nil {
		t.Fatal(err)
	}
	if code := get(t, srv, "/api/v1/tickets/"+tk.ID+"/land/preview", nil); code != 400 {
		t.Fatalf("preview no-repo = %d, want 400", code)
	}
}
