package sqlite

import (
	"errors"
	"testing"
	"time"

	"groundwork/internal/encoding"
	"groundwork/internal/ticket"
)

func TestContractIsCanonicalizedOnWrite(t *testing.T) {
	db := openTestDB(t)
	tk := &ticket.Ticket{Title: "node", NodeType: ticket.NodeComposite, Contract: `{"b":2,  "a":1}`}
	if err := db.CreateTicket(tk, "human"); err != nil {
		t.Fatal(err)
	}
	got, _ := db.GetTicket(tk.ID)
	if got.Contract != `{"a":1,"b":2}` {
		t.Errorf("contract = %q, want canonical {\"a\":1,\"b\":2}", got.Contract)
	}
}

func TestInvalidContractRejected(t *testing.T) {
	db := openTestDB(t)
	tk := &ticket.Ticket{Title: "node", Contract: `{not json`}
	if err := db.CreateTicket(tk, "human"); err == nil {
		t.Fatal("expected error for invalid contract JSON")
	}
}

func TestEmptyTitleRejected(t *testing.T) {
	db := openTestDB(t)
	if err := db.CreateTicket(&ticket.Ticket{Title: "  "}, "human"); !errors.Is(err, ErrEmptyTitle) {
		t.Fatalf("create: want ErrEmptyTitle, got %v", err)
	}
	tk := &ticket.Ticket{Title: "ok"}
	if err := db.CreateTicket(tk, "human"); err != nil {
		t.Fatal(err)
	}
	tk.Title = ""
	if err := db.UpdateTicket(tk, "human"); !errors.Is(err, ErrEmptyTitle) {
		t.Fatalf("update: want ErrEmptyTitle, got %v", err)
	}
}

func TestUpdateDoesNotChangeStatus(t *testing.T) {
	db := openTestDB(t)
	tk := &ticket.Ticket{Title: "node", Status: ticket.StatusTodo}
	if err := db.CreateTicket(tk, "human"); err != nil {
		t.Fatal(err)
	}
	// Attempt to sneak a status change through UpdateTicket.
	tk.Status = ticket.StatusDone
	tk.Title = "renamed"
	if err := db.UpdateTicket(tk, "human"); err != nil {
		t.Fatal(err)
	}
	got, _ := db.GetTicket(tk.ID)
	if got.Status != ticket.StatusTodo {
		t.Errorf("status = %q, want todo (UpdateTicket must not change status)", got.Status)
	}
	if got.Title != "renamed" {
		t.Errorf("title = %q, want renamed", got.Title)
	}
}

func TestRenewLeaseExpiredRejectedAndAudited(t *testing.T) {
	db := openTestDB(t)
	id := todoNode(t, db)
	if _, err := db.ClaimTicket(id, "run-1", "agent", time.Minute); err != nil {
		t.Fatal(err)
	}

	// A normal renew appends an audit event.
	if _, err := db.RenewLease(id, "run-1", time.Minute); err != nil {
		t.Fatal(err)
	}
	events, _ := db.AuditEventsFor("ticket", id)
	if events[len(events)-1].Type != "ticket.lease_renewed" {
		t.Errorf("last audit = %q, want ticket.lease_renewed", events[len(events)-1].Type)
	}

	// Force expiry, then renew must be rejected.
	past := encoding.FormatTime(time.Now().Add(-time.Hour))
	if _, err := db.Exec(`UPDATE leases SET expires_at=? WHERE ticket_id=?`, past, id); err != nil {
		t.Fatal(err)
	}
	if _, err := db.RenewLease(id, "run-1", time.Minute); !errors.Is(err, ErrLeaseExpired) {
		t.Fatalf("want ErrLeaseExpired, got %v", err)
	}
}
