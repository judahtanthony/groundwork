package server

import (
	"net/http"
	"strings"
	"testing"

	"groundwork/internal/ticket"
)

// The tickets page must classify todo work the same way the CLI does: a
// dependency-free todo node is ready, and one waiting on an unfinished node is
// blocked, annotated with the blocking dep and its status.
func TestTicketsPageReadyAndBlocked(t *testing.T) {
	srv, db := newTestServer(t)

	dep := &ticket.Ticket{Title: "Foundation", Status: ticket.StatusTodo, WorkType: "technical_implementation"}
	if err := db.CreateTicket(dep, "tester"); err != nil {
		t.Fatal(err)
	}
	waiting := &ticket.Ticket{Title: "Depends on foundation", Status: ticket.StatusTodo, WorkType: "technical_implementation"}
	if err := db.CreateTicket(waiting, "tester"); err != nil {
		t.Fatal(err)
	}
	if err := db.AddDependency(waiting.ID, dep.ID, "tester"); err != nil {
		t.Fatal(err)
	}

	rr := getHTML(t, srv, "/tickets")
	if rr.Code != http.StatusOK {
		t.Fatalf("GET /tickets = %d, want 200\n%s", rr.Code, rr.Body.String())
	}
	body := rr.Body.String()

	for _, want := range []string{"Ready", "Blocked", "All tickets", dep.ID, waiting.ID} {
		if !strings.Contains(body, want) {
			t.Errorf("tickets page missing %q", want)
		}
	}
	// The waiting node is annotated with its unmet dependency and that dep's status.
	if blocker := dep.ID + " (todo)"; !strings.Contains(body, blocker) {
		t.Errorf("tickets page missing blocker annotation %q", blocker)
	}
}
