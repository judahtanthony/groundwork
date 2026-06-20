package server

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"groundwork/internal/actor"
	"groundwork/internal/config"
	"groundwork/internal/eventbus"
	"groundwork/internal/policy"
	runpkg "groundwork/internal/run"
	"groundwork/internal/store/sqlite"
	"groundwork/internal/ticket"
)

const testActorsYAML = `schema: groundwork_actors/v1
actors:
  - id: human.owner
    type: human
    display_name: Owner
    roles: [owner]
  - id: ai.codex.default
    type: ai_agent
    display_name: Codex Default
    runtime: codex
`

// newTestServer builds a server over a migrated temp DB and a minimal project
// (with an actors.yaml) rooted in a temp dir.
func newTestServer(t *testing.T) (*Server, *sqlite.DB) {
	t.Helper()
	root := t.TempDir()
	gw := filepath.Join(root, config.GroundworkDir)
	if err := os.MkdirAll(gw, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(gw, "actors.yaml"), []byte(testActorsYAML), 0o644); err != nil {
		t.Fatal(err)
	}
	db, err := sqlite.Open(filepath.Join(gw, "state.sqlite"))
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	if err := db.Migrate(); err != nil {
		t.Fatalf("Migrate: %v", err)
	}
	cfg := config.Defaults()
	proj := &config.Project{Root: root, Config: &cfg}
	srv := New(db, proj, "test")
	reg, _, err := actor.Parse([]byte(testActorsYAML))
	if err != nil {
		t.Fatalf("parse actors: %v", err)
	}
	srv.SetApprovals(NewApprovalService(db, &policy.Set{}, reg))
	return srv, db
}

// get performs a GET against the server and decodes the JSON body into out
// (when non-nil), returning the status code.
func get(t *testing.T, srv *Server, path string, out any) int {
	t.Helper()
	rr := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rr, httptest.NewRequest("GET", path, nil))
	if out != nil && rr.Body.Len() > 0 {
		if err := json.Unmarshal(rr.Body.Bytes(), out); err != nil {
			t.Fatalf("decode %s: %v\nbody: %s", path, err, rr.Body.String())
		}
	}
	return rr.Code
}

// req performs a request with an optional JSON body and decodes the response.
func req(t *testing.T, srv *Server, method, path string, body, out any) int {
	t.Helper()
	var r *bytes.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			t.Fatal(err)
		}
		r = bytes.NewReader(b)
	} else {
		r = bytes.NewReader(nil)
	}
	rr := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rr, httptest.NewRequest(method, path, r))
	if out != nil && rr.Body.Len() > 0 {
		if err := json.Unmarshal(rr.Body.Bytes(), out); err != nil {
			t.Fatalf("decode %s %s: %v\nbody: %s", method, path, err, rr.Body.String())
		}
	}
	return rr.Code
}

func mustCreate(t *testing.T, db *sqlite.DB, tk *ticket.Ticket) *ticket.Ticket {
	t.Helper()
	if err := db.CreateTicket(tk, "test"); err != nil {
		t.Fatalf("create ticket: %v", err)
	}
	return tk
}

func TestHealthOK(t *testing.T) {
	srv, _ := newTestServer(t)
	var body map[string]string
	if code := get(t, srv, "/healthz", &body); code != http.StatusOK {
		t.Fatalf("status = %d, want 200", code)
	}
	if body["status"] != "ok" {
		t.Errorf("status field = %q, want ok", body["status"])
	}
}

func TestHealthStoreDown(t *testing.T) {
	srv, db := newTestServer(t)
	db.Close() // simulate an unavailable store
	var env errorEnvelope
	if code := get(t, srv, "/healthz", &env); code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want 503", code)
	}
	if env.Error.Code != "store_unavailable" {
		t.Errorf("error code = %q, want store_unavailable", env.Error.Code)
	}
}

func TestStateCounts(t *testing.T) {
	srv, db := newTestServer(t)
	mustCreate(t, db, &ticket.Ticket{Title: "a", Status: ticket.StatusTodo})
	mustCreate(t, db, &ticket.Ticket{Title: "b", Status: ticket.StatusTodo})
	mustCreate(t, db, &ticket.Ticket{Title: "c", Status: ticket.StatusDone})

	var got stateResponse
	if code := get(t, srv, "/api/v1/state", &got); code != http.StatusOK {
		t.Fatalf("status = %d, want 200", code)
	}
	if !got.OK || got.Version != "test" {
		t.Errorf("ok=%v version=%q, want true/\"test\"", got.OK, got.Version)
	}
	if got.Total != 3 {
		t.Errorf("total = %d, want 3", got.Total)
	}
	if got.Counts["todo"] != 2 || got.Counts["done"] != 1 {
		t.Errorf("counts = %v, want todo=2 done=1", got.Counts)
	}
	if got.Eligible != 2 { // two todo nodes, no dependencies
		t.Errorf("eligible = %d, want 2", got.Eligible)
	}
}

func TestTicketListAndGet(t *testing.T) {
	srv, db := newTestServer(t)
	mustCreate(t, db, &ticket.Ticket{Title: "a", Status: ticket.StatusTodo})

	var list []ticket.Ticket
	if code := get(t, srv, "/api/v1/tickets", &list); code != http.StatusOK {
		t.Fatalf("list status = %d, want 200", code)
	}
	if len(list) != 1 || list[0].ID != "T-0001" {
		t.Fatalf("list = %+v, want one T-0001", list)
	}

	var one ticket.Ticket
	if code := get(t, srv, "/api/v1/tickets/T-0001", &one); code != http.StatusOK {
		t.Fatalf("get status = %d, want 200", code)
	}
	if one.Title != "a" {
		t.Errorf("title = %q, want a", one.Title)
	}
}

func TestTicketGetNotFound(t *testing.T) {
	srv, _ := newTestServer(t)
	var env errorEnvelope
	if code := get(t, srv, "/api/v1/tickets/T-9999", &env); code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", code)
	}
	if env.Error.Code != "not_found" {
		t.Errorf("code = %q, want not_found", env.Error.Code)
	}
}

func TestTicketChildren(t *testing.T) {
	srv, db := newTestServer(t)
	parent := mustCreate(t, db, &ticket.Ticket{Title: "parent", Status: ticket.StatusTodo})
	mustCreate(t, db, &ticket.Ticket{Title: "child", Status: ticket.StatusBacklog, ParentID: parent.ID})

	var children []ticket.Ticket
	if code := get(t, srv, "/api/v1/tickets/"+parent.ID+"/children", &children); code != http.StatusOK {
		t.Fatalf("status = %d, want 200", code)
	}
	if len(children) != 1 || children[0].Title != "child" {
		t.Fatalf("children = %+v, want one child", children)
	}
}

func TestTicketDependencies(t *testing.T) {
	srv, db := newTestServer(t)
	a := mustCreate(t, db, &ticket.Ticket{Title: "a", Status: ticket.StatusTodo})
	b := mustCreate(t, db, &ticket.Ticket{Title: "b", Status: ticket.StatusTodo})
	if err := db.AddDependency(a.ID, b.ID, "test"); err != nil {
		t.Fatalf("add dependency: %v", err)
	}

	var deps dependenciesResponse
	if code := get(t, srv, "/api/v1/tickets/"+a.ID+"/dependencies", &deps); code != http.StatusOK {
		t.Fatalf("status = %d, want 200", code)
	}
	if len(deps.DependsOn) != 1 || deps.DependsOn[0] != b.ID {
		t.Errorf("depends_on = %v, want [%s]", deps.DependsOn, b.ID)
	}

	var depsB dependenciesResponse
	get(t, srv, "/api/v1/tickets/"+b.ID+"/dependencies", &depsB)
	if len(depsB.Dependents) != 1 || depsB.Dependents[0] != a.ID {
		t.Errorf("dependents = %v, want [%s]", depsB.Dependents, a.ID)
	}
}

func TestTicketContext(t *testing.T) {
	srv, db := newTestServer(t)
	parent := mustCreate(t, db, &ticket.Ticket{Title: "parent", Status: ticket.StatusInProgress})
	child := mustCreate(t, db, &ticket.Ticket{Title: "child", Status: ticket.StatusTodo, ParentID: parent.ID})

	var brief struct {
		Node          struct{ ID string } `json:"node"`
		AncestorSpine []struct {
			ID string `json:"id"`
		} `json:"ancestor_spine"`
	}
	if code := get(t, srv, "/api/v1/tickets/"+child.ID+"/context", &brief); code != http.StatusOK {
		t.Fatalf("status = %d, want 200", code)
	}
	if brief.Node.ID != child.ID {
		t.Errorf("node id = %q, want %s", brief.Node.ID, child.ID)
	}
	if len(brief.AncestorSpine) != 1 || brief.AncestorSpine[0].ID != parent.ID {
		t.Errorf("ancestor spine = %+v, want [%s]", brief.AncestorSpine, parent.ID)
	}
}

func TestActorListAndGet(t *testing.T) {
	srv, _ := newTestServer(t)

	var actors []struct {
		ID   string `json:"id"`
		Type string `json:"type"`
	}
	if code := get(t, srv, "/api/v1/actors", &actors); code != http.StatusOK {
		t.Fatalf("status = %d, want 200", code)
	}
	found := false
	for _, a := range actors {
		if a.ID == "ai.codex.default" {
			found = true
		}
	}
	if !found {
		t.Errorf("actors = %+v, want ai.codex.default present", actors)
	}

	var one struct {
		ID string `json:"id"`
	}
	if code := get(t, srv, "/api/v1/actors/human.owner", &one); code != http.StatusOK {
		t.Fatalf("actor get status = %d, want 200", code)
	}
	if one.ID != "human.owner" {
		t.Errorf("actor id = %q, want human.owner", one.ID)
	}

	var env errorEnvelope
	if code := get(t, srv, "/api/v1/actors/nope", &env); code != http.StatusNotFound {
		t.Errorf("missing actor status = %d, want 404", code)
	}
}

func TestTicketCreateAndPatch(t *testing.T) {
	srv, _ := newTestServer(t)

	var created ticket.Ticket
	if code := req(t, srv, "POST", "/api/v1/tickets",
		map[string]any{"title": "new", "status": "todo"}, &created); code != http.StatusCreated {
		t.Fatalf("create status = %d, want 201", code)
	}
	if created.ID == "" || created.Title != "new" {
		t.Fatalf("created = %+v, want id + title", created)
	}

	var patched ticket.Ticket
	if code := req(t, srv, "PATCH", "/api/v1/tickets/"+created.ID,
		map[string]any{"title": "renamed", "kind": "ticket"}, &patched); code != http.StatusOK {
		t.Fatalf("patch status = %d, want 200", code)
	}
	if patched.Title != "renamed" {
		t.Errorf("patched title = %q, want renamed", patched.Title)
	}
}

func TestTicketCreateEmptyTitle(t *testing.T) {
	srv, _ := newTestServer(t)
	var env errorEnvelope
	if code := req(t, srv, "POST", "/api/v1/tickets", map[string]any{"status": "todo"}, &env); code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", code)
	}
	if env.Error.Code != "empty_title" {
		t.Errorf("code = %q, want empty_title", env.Error.Code)
	}
}

func TestTicketTransition(t *testing.T) {
	srv, db := newTestServer(t)
	tk := mustCreate(t, db, &ticket.Ticket{Title: "a", Status: ticket.StatusTodo})

	if code := req(t, srv, "POST", "/api/v1/tickets/"+tk.ID+"/transition",
		map[string]string{"status": "in_progress"}, nil); code != http.StatusOK {
		t.Fatalf("transition status = %d, want 200", code)
	}

	// todo -> done is not a legal direct edge: expect 409 illegal_transition.
	tk2 := mustCreate(t, db, &ticket.Ticket{Title: "b", Status: ticket.StatusTodo})
	var env errorEnvelope
	if code := req(t, srv, "POST", "/api/v1/tickets/"+tk2.ID+"/transition",
		map[string]string{"status": "done"}, &env); code != http.StatusConflict {
		t.Fatalf("illegal transition status = %d, want 409", code)
	}
	if env.Error.Code != "illegal_transition" {
		t.Errorf("code = %q, want illegal_transition", env.Error.Code)
	}
}

func TestTicketAddRemoveDependency(t *testing.T) {
	srv, db := newTestServer(t)
	a := mustCreate(t, db, &ticket.Ticket{Title: "a", Status: ticket.StatusTodo})
	b := mustCreate(t, db, &ticket.Ticket{Title: "b", Status: ticket.StatusTodo})

	if code := req(t, srv, "POST", "/api/v1/tickets/"+a.ID+"/dependencies",
		map[string]string{"depends_on": b.ID}, nil); code != http.StatusOK {
		t.Fatalf("add dependency status = %d, want 200", code)
	}

	// Reverse edge would create a cycle -> 409.
	var env errorEnvelope
	if code := req(t, srv, "POST", "/api/v1/tickets/"+b.ID+"/dependencies",
		map[string]string{"depends_on": a.ID}, &env); code != http.StatusConflict {
		t.Fatalf("cycle status = %d, want 409", code)
	}
	if env.Error.Code != "dependency_cycle" {
		t.Errorf("code = %q, want dependency_cycle", env.Error.Code)
	}

	if code := req(t, srv, "DELETE", "/api/v1/tickets/"+a.ID+"/dependencies/"+b.ID, nil, nil); code != http.StatusOK {
		t.Fatalf("remove dependency status = %d, want 200", code)
	}
	var deps dependenciesResponse
	get(t, srv, "/api/v1/tickets/"+a.ID+"/dependencies", &deps)
	if len(deps.DependsOn) != 0 {
		t.Errorf("depends_on = %v, want empty after removal", deps.DependsOn)
	}
}

func startRun(t *testing.T, db *sqlite.DB, title string) *sqlite.Run {
	t.Helper()
	tk := &ticket.Ticket{Title: title, Status: ticket.StatusTodo}
	if err := db.CreateTicket(tk, "tester"); err != nil {
		t.Fatal(err)
	}
	r, _, err := db.StartRun(sqlite.StartRunParams{
		TicketID: tk.ID, ActorID: "ai.codex.default", Mode: runpkg.ModeImplementation,
		Runtime: "stub", TTL: 90 * time.Second,
	})
	if err != nil {
		t.Fatalf("StartRun: %v", err)
	}
	return r
}

func TestRunReadEndpoints(t *testing.T) {
	srv, db := newTestServer(t)
	r := startRun(t, db, "worked")
	if _, err := db.AppendRunEvent(r.ID, "working", "msg", nil); err != nil {
		t.Fatal(err)
	}

	var list []sqlite.Run
	if code := get(t, srv, "/api/v1/runs", &list); code != http.StatusOK || len(list) != 1 {
		t.Fatalf("list code=%d len=%d", code, len(list))
	}
	var one sqlite.Run
	if code := get(t, srv, "/api/v1/runs/"+r.ID, &one); code != http.StatusOK || one.ID != r.ID {
		t.Fatalf("get code=%d run=%+v", code, one)
	}
	var events []sqlite.RunEvent
	if code := get(t, srv, "/api/v1/runs/"+r.ID+"/events", &events); code != http.StatusOK || len(events) != 1 {
		t.Fatalf("events code=%d len=%d", code, len(events))
	}
}

func TestRunPauseResumeCancel(t *testing.T) {
	srv, db := newTestServer(t)
	r := startRun(t, db, "lifecycle")

	var paused sqlite.Run
	if code := req(t, srv, "POST", "/api/v1/runs/"+r.ID+"/pause", nil, &paused); code != http.StatusOK {
		t.Fatalf("pause code=%d", code)
	}
	if paused.Status != string(runpkg.StatusPaused) {
		t.Errorf("status = %s, want paused", paused.Status)
	}
	if code := req(t, srv, "POST", "/api/v1/runs/"+r.ID+"/resume", nil, nil); code != http.StatusOK {
		t.Fatalf("resume code=%d", code)
	}
	var cancelled sqlite.Run
	if code := req(t, srv, "POST", "/api/v1/runs/"+r.ID+"/cancel", nil, &cancelled); code != http.StatusOK {
		t.Fatalf("cancel code=%d", code)
	}
	if cancelled.Status != string(runpkg.StatusCancelled) {
		t.Errorf("status = %s, want cancelled", cancelled.Status)
	}
	// Cancel released the lease and returned the node to blocked.
	tk, _ := db.GetTicket(r.TicketID)
	if tk.Status != ticket.StatusBlocked {
		t.Errorf("ticket status = %s, want blocked", tk.Status)
	}
	if lease, _ := db.GetLease(r.TicketID); lease != nil {
		t.Errorf("lease still present after cancel: %+v", lease)
	}
}

// fakeDispatcher is a stand-in scheduler for the run-trigger endpoints.
type fakeDispatcher struct{ run *sqlite.Run }

func (f *fakeDispatcher) RunOnce(_ context.Context, ticketID string) (*sqlite.Run, error) {
	return f.run, nil
}
func (f *fakeDispatcher) Tick(_ context.Context) (int, error) { return 3, nil }

func TestRunCreateNeedsScheduler(t *testing.T) {
	srv, _ := newTestServer(t)
	if code := req(t, srv, "POST", "/api/v1/runs", map[string]string{}, nil); code != http.StatusServiceUnavailable {
		t.Fatalf("no-scheduler code = %d, want 503", code)
	}
}

func TestRunCreateOnceAndNext(t *testing.T) {
	srv, db := newTestServer(t)
	r := startRun(t, db, "triggered")
	srv.SetScheduler(&fakeDispatcher{run: r})

	var once sqlite.Run
	if code := req(t, srv, "POST", "/api/v1/runs", map[string]string{"ticket_id": r.TicketID}, &once); code != http.StatusCreated {
		t.Fatalf("run once code=%d", code)
	}
	var next struct {
		Started int `json:"started"`
	}
	if code := req(t, srv, "POST", "/api/v1/runs", map[string]string{}, &next); code != http.StatusOK || next.Started != 3 {
		t.Fatalf("run next code=%d started=%d", code, next.Started)
	}
}

func TestDecomposeApproveFlow(t *testing.T) {
	srv, db := newTestServer(t)
	parent := &ticket.Ticket{Title: "epic", Status: ticket.StatusInProgress}
	if err := db.CreateTicket(parent, "tester"); err != nil {
		t.Fatal(err)
	}
	if err := db.TriageTicket(parent.ID, ticket.NodeComposite, "tester"); err != nil {
		t.Fatal(err)
	}

	var proposal struct {
		Approval *sqlite.Approval `json:"approval"`
		ChildIDs []string         `json:"child_ids"`
	}
	body := map[string]any{
		"contract": json.RawMessage(`{"schema":"c/v1"}`),
		"children": []map[string]string{{"title": "child one"}},
	}
	if code := req(t, srv, "POST", "/api/v1/tickets/"+parent.ID+"/decompose", body, &proposal); code != http.StatusCreated {
		t.Fatalf("decompose code=%d", code)
	}
	if len(proposal.ChildIDs) != 1 || proposal.Approval.Type != "decompose" {
		t.Fatalf("proposal=%+v", proposal)
	}

	// Approve the proposal -> child becomes todo.
	if code := req(t, srv, "POST", "/api/v1/approvals/"+proposal.Approval.ID+"/approve", map[string]string{"reason": "ok"}, nil); code != http.StatusOK {
		t.Fatalf("approve code=%d", code)
	}
	child, _ := db.GetTicket(proposal.ChildIDs[0])
	if child.Status != ticket.StatusTodo {
		t.Errorf("child status = %s, want todo", child.Status)
	}

	pending, _ := db.ListApprovals("pending")
	if len(pending) != 0 {
		t.Errorf("pending approvals = %d, want 0", len(pending))
	}
}

func TestEscalateReplanFlow(t *testing.T) {
	srv, db := newTestServer(t)
	tk := &ticket.Ticket{Title: "node", Status: ticket.StatusInProgress}
	if err := db.CreateTicket(tk, "tester"); err != nil {
		t.Fatal(err)
	}

	var appr sqlite.Approval
	if code := req(t, srv, "POST", "/api/v1/tickets/"+tk.ID+"/escalate", map[string]string{"reason": "changed"}, &appr); code != http.StatusCreated {
		t.Fatalf("escalate code=%d", code)
	}
	if appr.Type != "replan" {
		t.Fatalf("type = %s, want replan", appr.Type)
	}
	if got, _ := db.GetTicket(tk.ID); got.Status != ticket.StatusBlocked {
		t.Errorf("ticket status = %s, want blocked", got.Status)
	}

	if code := req(t, srv, "POST", "/api/v1/approvals/"+appr.ID+"/approve", nil, nil); code != http.StatusOK {
		t.Fatalf("approve replan code=%d", code)
	}
	if got, _ := db.GetTicket(tk.ID); got.Status != ticket.StatusTodo {
		t.Errorf("ticket status = %s, want todo", got.Status)
	}
}

func TestLandThroughGate(t *testing.T) {
	srv, db := newTestServer(t)
	tk := &ticket.Ticket{Title: "feature", Status: ticket.StatusReview}
	if err := db.CreateTicket(tk, "tester"); err != nil {
		t.Fatal(err)
	}

	// Landing opens a human approval (no auto-approve rule, empty Scope).
	var pending landResponse
	if code := req(t, srv, "POST", "/api/v1/tickets/"+tk.ID+"/land", map[string]bool{}, &pending); code != http.StatusOK {
		t.Fatalf("land request code = %d, want 200", code)
	}
	if pending.Landed || pending.Approval == nil || pending.Approval.Type != "land_to_main" {
		t.Fatalf("expected pending land_to_main approval, got %+v", pending)
	}
	if got, _ := db.GetTicket(tk.ID); got.Status != ticket.StatusReview {
		t.Errorf("node landed before approval: %s", got.Status)
	}

	// Approving the land_to_main approval performs the land.
	if code := req(t, srv, "POST", "/api/v1/approvals/"+pending.Approval.ID+"/approve", nil, nil); code != http.StatusOK {
		t.Fatalf("approve land code = %d, want 200", code)
	}
	if got, _ := db.GetTicket(tk.ID); got.Status != ticket.StatusDone {
		t.Errorf("status = %s, want done", got.Status)
	}
}

func TestLandValidationGateAndOverride(t *testing.T) {
	srv, db := newTestServer(t)
	tk := &ticket.Ticket{Title: "feature", Status: ticket.StatusReview}
	if err := db.CreateTicket(tk, "tester"); err != nil {
		t.Fatal(err)
	}
	if _, err := db.RecordValidation(sqlite.ValidationResult{TicketID: tk.ID, Name: "go_tests", Status: sqlite.ValidationFail}); err != nil {
		t.Fatal(err)
	}

	// Open the land approval, then approving it is blocked by the failing
	// validation (409); the approval stays pending.
	var pending landResponse
	req(t, srv, "POST", "/api/v1/tickets/"+tk.ID+"/land", map[string]bool{}, &pending)
	if pending.Approval == nil {
		t.Fatal("expected a pending approval")
	}
	if code := req(t, srv, "POST", "/api/v1/approvals/"+pending.Approval.ID+"/approve", nil, nil); code != http.StatusConflict {
		t.Fatalf("approve-with-failing-validation code = %d, want 409", code)
	}
	if got, _ := db.GetApproval(pending.Approval.ID); got.Status != "pending" {
		t.Errorf("approval status = %s, want pending (land blocked)", got.Status)
	}

	// Validations are listable.
	var results []sqlite.ValidationResult
	if code := get(t, srv, "/api/v1/tickets/"+tk.ID+"/validations", &results); code != http.StatusOK || len(results) != 1 {
		t.Fatalf("validations code=%d len=%d", code, len(results))
	}

	// Override bypasses both gates and lands.
	var landed landResponse
	if code := req(t, srv, "POST", "/api/v1/tickets/"+tk.ID+"/land", map[string]bool{"override": true}, &landed); code != http.StatusOK {
		t.Fatalf("override land code = %d, want 200", code)
	}
	if !landed.Landed || landed.Ticket.Status != ticket.StatusDone {
		t.Errorf("override result = %+v, want landed done", landed)
	}
}

func TestSSEStreamsEvents(t *testing.T) {
	srv, _ := newTestServer(t)
	bus := eventbus.New(8)
	srv.SetBus(bus)
	ts := httptest.NewServer(srv.Handler())
	defer ts.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	req, _ := http.NewRequestWithContext(ctx, "GET", ts.URL+"/api/v1/events", nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	if ct := resp.Header.Get("Content-Type"); ct != "text/event-stream" {
		t.Fatalf("content-type = %q", ct)
	}

	reader := bufio.NewReader(resp.Body)
	// Wait for the connection comment so we know the handler has subscribed.
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			t.Fatalf("reading connect frame: %v", err)
		}
		if strings.Contains(line, "connected") {
			break
		}
	}

	bus.Publish(eventbus.Event{Type: "run.started", RunID: "R-0001"})

	found := false
	for i := 0; i < 20 && !found; i++ {
		line, err := reader.ReadString('\n')
		if err != nil {
			break
		}
		if strings.Contains(line, "run.started") {
			found = true
		}
	}
	if !found {
		t.Fatal("did not receive published event over SSE")
	}
}

func TestRecordValidationEndpoint(t *testing.T) {
	srv, db := newTestServer(t)
	tk := &ticket.Ticket{Title: "node", Status: ticket.StatusInProgress}
	if err := db.CreateTicket(tk, "tester"); err != nil {
		t.Fatal(err)
	}
	var rec sqlite.ValidationResult
	if code := req(t, srv, "POST", "/api/v1/tickets/"+tk.ID+"/validations",
		map[string]string{"name": "go_tests", "command": "go test ./...", "status": "pass"}, &rec); code != http.StatusCreated {
		t.Fatalf("record code = %d, want 201", code)
	}
	if rec.ID == "" || rec.Status != "pass" {
		t.Fatalf("recorded = %+v", rec)
	}
	var list []sqlite.ValidationResult
	get(t, srv, "/api/v1/tickets/"+tk.ID+"/validations", &list)
	if len(list) != 1 {
		t.Errorf("validations = %d, want 1", len(list))
	}
}

func TestUnknownRoute404(t *testing.T) {
	srv, _ := newTestServer(t)
	if code := get(t, srv, "/api/v1/nope", nil); code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", code)
	}
}
