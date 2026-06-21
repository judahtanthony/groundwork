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

// landToPending opens a landing and returns the pending approval id. Fails if the
// landing did not produce a human gate.
func landToPending(t *testing.T, srv *Server, id string) string {
	t.Helper()
	var pending landResponse
	if code := req(t, srv, "POST", "/api/v1/tickets/"+id+"/land", map[string]bool{}, &pending); code != http.StatusOK {
		t.Fatalf("land code = %d", code)
	}
	if pending.Approval == nil {
		t.Fatalf("expected a pending approval, got %+v", pending)
	}
	return pending.Approval.ID
}

// TestLandCommitFailureIsRecoverable asserts that when the git commit fails, the
// node is recorded done-but-uncommitted with a clear error, and re-running land
// finishes the commit (ADR 0034 / CR#1).
func TestLandCommitFailureIsRecoverable(t *testing.T) {
	srv, db, root := newGitServer(t)
	tk := &ticket.Ticket{Title: "doc", Status: ticket.StatusReview, WorkType: "documentation"}
	if err := db.CreateTicket(tk, "tester"); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(root, "docs"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "docs", "note.md"), []byte("# Note\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	runGit(t, root, "add", "docs/note.md")

	// A pre-commit hook that rejects every commit.
	hook := filepath.Join(root, ".git", "hooks", "pre-commit")
	if err := os.WriteFile(hook, []byte("#!/bin/sh\nexit 1\n"), 0o755); err != nil {
		t.Fatal(err)
	}

	before := runGit(t, root, "rev-parse", "HEAD")
	appr := landToPending(t, srv, tk.ID)

	var env errorEnvelope
	if code := req(t, srv, "POST", "/api/v1/approvals/"+appr+"/approve", nil, &env); code != http.StatusInternalServerError {
		t.Fatalf("approve code = %d, want 500", code)
	}
	if env.Error.Code != "land_commit_failed" || !strings.Contains(env.Error.Message, "gw ticket land "+tk.ID) {
		t.Errorf("error = %+v, want land_commit_failed naming the recovery command", env.Error)
	}
	// Node is recorded done; nothing committed yet.
	if got, _ := db.GetTicket(tk.ID); got.Status != ticket.StatusDone {
		t.Fatalf("status = %s, want done (recorded landed)", got.Status)
	}
	if runGit(t, root, "rev-parse", "HEAD") != before {
		t.Fatal("HEAD advanced despite the failed commit")
	}

	// Fix the environment and re-run land: the already-done node finishes the commit.
	if err := os.Remove(hook); err != nil {
		t.Fatal(err)
	}
	var landed landResponse
	if code := req(t, srv, "POST", "/api/v1/tickets/"+tk.ID+"/land", map[string]bool{}, &landed); code != http.StatusOK {
		t.Fatalf("recovery land code = %d, want 200", code)
	}
	after := runGit(t, root, "rev-parse", "HEAD")
	if after == before {
		t.Fatal("recovery did not produce a commit")
	}
	files := runGit(t, root, "show", "--name-only", "--pretty=format:", after)
	for _, want := range []string{"docs/note.md", filepath.Join(config.GroundworkDir, "tickets", tk.ID, "ticket.md")} {
		if !strings.Contains(files, want) {
			t.Errorf("recovery commit missing %s; files:\n%s", want, files)
		}
	}
}

// TestLandRefusesDetachedHead asserts a landing on a detached HEAD is refused
// rather than producing an orphan commit (CR#8).
func TestLandRefusesDetachedHead(t *testing.T) {
	srv, db, root := newGitServer(t)
	runGit(t, root, "checkout", "--detach")

	tk := &ticket.Ticket{Title: "doc", Status: ticket.StatusReview, WorkType: "documentation"}
	if err := db.CreateTicket(tk, "tester"); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "a.md"), []byte("a\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	runGit(t, root, "add", "a.md")

	before := runGit(t, root, "rev-parse", "HEAD")
	appr := landToPending(t, srv, tk.ID)
	var env errorEnvelope
	if code := req(t, srv, "POST", "/api/v1/approvals/"+appr+"/approve", nil, &env); code != http.StatusInternalServerError {
		t.Fatalf("approve code = %d, want 500", code)
	}
	if !strings.Contains(env.Error.Message, "detached") {
		t.Errorf("error = %q, want a detached-HEAD message", env.Error.Message)
	}
	if runGit(t, root, "rev-parse", "HEAD") != before {
		t.Fatal("a commit was made on a detached HEAD")
	}
}

// TestLandCommitsMultipleFiles asserts the ticket-scoped pathspec can be more than
// one staged file (CR test gap).
func TestLandCommitsMultipleFiles(t *testing.T) {
	srv, db, root := newGitServer(t)
	tk := &ticket.Ticket{Title: "doc", Status: ticket.StatusReview, WorkType: "documentation"}
	if err := db.CreateTicket(tk, "tester"); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(root, "docs"), 0o755); err != nil {
		t.Fatal(err)
	}
	for _, f := range []string{"a.md", "b.md"} {
		if err := os.WriteFile(filepath.Join(root, "docs", f), []byte("x\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	runGit(t, root, "add", "docs/a.md", "docs/b.md")

	appr := landToPending(t, srv, tk.ID)
	if code := req(t, srv, "POST", "/api/v1/approvals/"+appr+"/approve", nil, nil); code != http.StatusOK {
		t.Fatalf("approve code = %d", code)
	}
	head := runGit(t, root, "rev-parse", "HEAD")
	files := runGit(t, root, "show", "--name-only", "--pretty=format:", head)
	for _, want := range []string{"docs/a.md", "docs/b.md", filepath.Join(config.GroundworkDir, "tickets", tk.ID, "ticket.md")} {
		if !strings.Contains(files, want) {
			t.Errorf("commit missing %s; files:\n%s", want, files)
		}
	}
}
