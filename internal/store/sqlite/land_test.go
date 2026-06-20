package sqlite

import (
	"errors"
	"testing"

	"groundwork/internal/ticket"
)

// approvedTicket creates a ticket already in the approved state.
func approvedTicket(t *testing.T, db *DB, title string) *ticket.Ticket {
	t.Helper()
	tk := &ticket.Ticket{Title: title, Status: ticket.StatusApproved}
	if err := db.CreateTicket(tk, "tester"); err != nil {
		t.Fatal(err)
	}
	return tk
}

func TestLandRequiresApproved(t *testing.T) {
	db := openTestDB(t)
	tk := todoTicket(t, db, "not approved")
	if _, err := db.Land(tk.ID, nil, false, "human.owner"); !errors.Is(err, ErrNotApproved) {
		t.Fatalf("err = %v, want ErrNotApproved", err)
	}
}

func TestLandBlockedByMissingRequiredCheck(t *testing.T) {
	db := openTestDB(t)
	tk := approvedTicket(t, db, "needs go tests")
	lr, err := db.Land(tk.ID, []string{"go_tests"}, false, "human.owner")
	if !errors.Is(err, ErrValidationGate) {
		t.Fatalf("err = %v, want ErrValidationGate", err)
	}
	if lr == nil || len(lr.Missing) != 1 {
		t.Fatalf("land result = %+v, want missing go_tests", lr)
	}
	if got, _ := db.GetTicket(tk.ID); got.Status != ticket.StatusApproved {
		t.Errorf("status = %s, want approved (not landed)", got.Status)
	}
}

func TestLandBlockedByFailingResult(t *testing.T) {
	db := openTestDB(t)
	tk := approvedTicket(t, db, "failing")
	if _, err := db.RecordValidation(ValidationResult{TicketID: tk.ID, Name: "go_tests", Status: ValidationFail}); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Land(tk.ID, nil, false, "human.owner"); !errors.Is(err, ErrValidationGate) {
		t.Fatalf("err = %v, want ErrValidationGate (failing result)", err)
	}
}

func TestLandSucceedsWithPassingChecks(t *testing.T) {
	db := openTestDB(t)
	tk := approvedTicket(t, db, "passing")
	if _, err := db.RecordValidation(ValidationResult{TicketID: tk.ID, Name: "go_tests", Status: ValidationPass}); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Land(tk.ID, []string{"go_tests"}, false, "human.owner"); err != nil {
		t.Fatalf("Land: %v", err)
	}
	if got, _ := db.GetTicket(tk.ID); got.Status != ticket.StatusDone {
		t.Errorf("status = %s, want done", got.Status)
	}
}

func TestLandOverrideBypassesGate(t *testing.T) {
	db := openTestDB(t)
	tk := approvedTicket(t, db, "override")
	if _, err := db.RecordValidation(ValidationResult{TicketID: tk.ID, Name: "go_tests", Status: ValidationFail}); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Land(tk.ID, []string{"go_tests"}, true, "human.owner"); err != nil {
		t.Fatalf("override Land: %v", err)
	}
	if got, _ := db.GetTicket(tk.ID); got.Status != ticket.StatusDone {
		t.Errorf("status = %s, want done after override", got.Status)
	}
}
