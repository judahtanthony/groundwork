package server

import (
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"groundwork/internal/actor"
	"groundwork/internal/config"
	"groundwork/internal/policy"
	"groundwork/internal/store/sqlite"
	"groundwork/internal/ticket"
)

func runGit(t *testing.T, dir string, args ...string) string {
	t.Helper()
	c := exec.Command("git", args...)
	c.Dir = dir
	out, err := c.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v: %s", args, err, out)
	}
	return strings.TrimSpace(string(out))
}

// newGitServer builds a server whose project root is a git work tree, so
// landings commit there (ADR 0034 / T-1004).
func newGitServer(t *testing.T) (*Server, *sqlite.DB, string) {
	t.Helper()
	root := t.TempDir()
	runGit(t, root, "init")
	runGit(t, root, "config", "user.email", "t@example.com")
	runGit(t, root, "config", "user.name", "Test")
	runGit(t, root, "commit", "--allow-empty", "-m", "init") // a real repo always has history

	gw := filepath.Join(root, config.GroundworkDir)
	if err := os.MkdirAll(gw, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(gw, "actors.yaml"), []byte(testActorsYAML), 0o644); err != nil {
		t.Fatal(err)
	}
	db, err := sqlite.Open(filepath.Join(gw, "state.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	if err := db.Migrate(); err != nil {
		t.Fatal(err)
	}
	cfg := config.Defaults()
	proj := &config.Project{Root: root, Config: &cfg}
	srv := New(db, proj, "test")
	if srv.repo == nil {
		t.Fatal("git repo not detected at project root")
	}
	reg, _, err := actor.Parse([]byte(testActorsYAML))
	if err != nil {
		t.Fatal(err)
	}
	srv.SetApprovals(NewApprovalService(db, &policy.Set{}, reg))
	return srv, db, root
}

// TestLandCommitsToGit drives a docs node through the human landing gate and
// asserts gw makes the durable commit: HEAD advances, the commit carries both the
// human's staged doc and the regenerated export (status: done), and the SHA is on
// the audit trail.
func TestLandCommitsToGit(t *testing.T) {
	srv, db, root := newGitServer(t)
	tk := &ticket.Ticket{Title: "document the thing", Status: ticket.StatusReview, WorkType: "documentation"}
	if err := db.CreateTicket(tk, "tester"); err != nil {
		t.Fatal(err)
	}

	// The human edits a doc and stages it (the ticket-scoped pathspec).
	if err := os.MkdirAll(filepath.Join(root, "docs"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "docs", "note.md"), []byte("# Note\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	runGit(t, root, "add", "docs/note.md")

	before := runGit(t, root, "rev-parse", "HEAD")

	var pending landResponse
	if code := req(t, srv, "POST", "/api/v1/tickets/"+tk.ID+"/land", map[string]bool{}, &pending); code != http.StatusOK {
		t.Fatalf("land code = %d", code)
	}
	if pending.Landed || pending.Approval == nil {
		t.Fatalf("expected pending approval, got %+v", pending)
	}
	if code := req(t, srv, "POST", "/api/v1/approvals/"+pending.Approval.ID+"/approve", nil, nil); code != http.StatusOK {
		t.Fatalf("approve code = %d", code)
	}

	if got, _ := db.GetTicket(tk.ID); got.Status != ticket.StatusDone {
		t.Fatalf("status = %s, want done", got.Status)
	}

	after := runGit(t, root, "rev-parse", "HEAD")
	if after == before {
		t.Fatal("HEAD did not advance: no landing commit")
	}

	// The commit contains the human's doc and the regenerated export.
	files := runGit(t, root, "show", "--name-only", "--pretty=format:", after)
	exportRel := filepath.Join(config.GroundworkDir, "tickets", tk.ID, "ticket.md")
	for _, want := range []string{"docs/note.md", exportRel} {
		if !strings.Contains(files, want) {
			t.Errorf("commit missing %s; files:\n%s", want, files)
		}
	}

	// The committed export shows status: done.
	data, err := os.ReadFile(filepath.Join(root, exportRel))
	if err != nil || !strings.Contains(string(data), "status: done") {
		t.Errorf("export missing status: done (err %v):\n%s", err, data)
	}

	// The commit SHA is recorded on the audit trail.
	events, _ := db.AuditEventsFor("ticket", tk.ID)
	var committed bool
	for _, e := range events {
		if e.Type == "ticket.committed" && strings.Contains(e.Payload, after) {
			committed = true
		}
	}
	if !committed {
		t.Errorf("no ticket.committed audit event with sha %s: %+v", after, events)
	}
}

// TestLandWithoutGitStillRecords confirms graceful degradation: a non-git
// project still lands (status done) without a commit.
func TestLandWithoutGitStillRecords(t *testing.T) {
	srv, db := newTestServer(t)
	if srv.repo != nil {
		t.Fatal("did not expect a git repo in a bare temp dir")
	}
	tk := &ticket.Ticket{Title: "docs", Status: ticket.StatusReview}
	if err := db.CreateTicket(tk, "tester"); err != nil {
		t.Fatal(err)
	}
	var pending landResponse
	if code := req(t, srv, "POST", "/api/v1/tickets/"+tk.ID+"/land", map[string]bool{}, &pending); code != http.StatusOK {
		t.Fatalf("land code = %d", code)
	}
	if code := req(t, srv, "POST", "/api/v1/approvals/"+pending.Approval.ID+"/approve", nil, nil); code != http.StatusOK {
		t.Fatalf("approve code = %d", code)
	}
	if got, _ := db.GetTicket(tk.ID); got.Status != ticket.StatusDone {
		t.Fatalf("status = %s, want done", got.Status)
	}
}
