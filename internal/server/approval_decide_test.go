package server

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"groundwork/internal/approval"
	"groundwork/internal/store/sqlite"
	"groundwork/internal/ticket"
)

func postForm(t *testing.T, srv *Server, path string, form url.Values) *httptest.ResponseRecorder {
	t.Helper()
	rr := httptest.NewRecorder()
	r := httptest.NewRequest("POST", path, strings.NewReader(form.Encode()))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	srv.Handler().ServeHTTP(rr, r)
	return rr
}

// pendingApproval seeds a ticket and a pending land_to_main approval, returning
// the approval id.
func pendingApproval(t *testing.T, db *sqlite.DB) string {
	t.Helper()
	tk := &ticket.Ticket{Title: "Decide me", Status: ticket.StatusInProgress, WorkType: "technical_implementation"}
	if err := db.CreateTicket(tk, "tester"); err != nil {
		t.Fatal(err)
	}
	a, err := db.CreateApproval(sqlite.CreateApprovalParams{
		TicketID: tk.ID, Type: approval.TypeLandToMain, RiskClass: "high",
		Summary: "land it", Status: approval.StatusPending, RequestedByActor: "ai.codex.default",
	})
	if err != nil {
		t.Fatal(err)
	}
	return a.ID
}

// A reject from the inbox routes through the ApprovalService and redirects back to
// the inbox, leaving the approval rejected (not pending).
func TestApprovalDecideRejectRedirects(t *testing.T) {
	srv, db := newTestServer(t)
	id := pendingApproval(t, db)

	rr := postForm(t, srv, "/approvals/"+id+"/decide", url.Values{
		"decision": {"reject"}, "reason": {"not yet"},
	})
	if rr.Code != http.StatusSeeOther {
		t.Fatalf("decide reject = %d, want 303\n%s", rr.Code, rr.Body.String())
	}
	if loc := rr.Header().Get("Location"); loc != "/approvals" {
		t.Errorf("redirect Location = %q, want /approvals", loc)
	}
	a, err := db.GetApproval(id)
	if err != nil {
		t.Fatal(err)
	}
	if a.Status != string(approval.StatusRejected) {
		t.Errorf("approval status = %q, want rejected", a.Status)
	}
	if a.DecisionReason != "not yet" {
		t.Errorf("decision reason = %q, want %q", a.DecisionReason, "not yet")
	}
}

// An unknown decision value is rejected before touching the service.
func TestApprovalDecideRejectsBadDecision(t *testing.T) {
	srv, db := newTestServer(t)
	id := pendingApproval(t, db)
	rr := postForm(t, srv, "/approvals/"+id+"/decide", url.Values{"decision": {"explode"}})
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("bad decision = %d, want 400", rr.Code)
	}
}

// The inbox renders the decision form so the actions are reachable.
func TestApprovalsInboxRendersDecisionForm(t *testing.T) {
	srv, db := newTestServer(t)
	id := pendingApproval(t, db)
	body := getHTML(t, srv, "/approvals").Body.String()
	for _, want := range []string{
		`action="/approvals/` + id + `/decide"`,
		`name="decision" value="approve"`,
		`name="decision" value="reject"`,
		`name="decision" value="clarify"`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("inbox missing decision form element %q", want)
		}
	}
}
