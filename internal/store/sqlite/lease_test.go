package sqlite

import (
	"errors"
	"sync"
	"testing"
	"time"

	"groundwork/internal/encoding"
	"groundwork/internal/ticket"
)

func todoNode(t *testing.T, db *DB) string {
	t.Helper()
	tk := &ticket.Ticket{Title: "node", Status: ticket.StatusTodo}
	if err := db.CreateTicket(tk, "human"); err != nil {
		t.Fatal(err)
	}
	return tk.ID
}

func TestClaimSingleWinner(t *testing.T) {
	db := openTestDB(t)
	id := todoNode(t, db)

	const n = 8
	var wg sync.WaitGroup
	var mu sync.Mutex
	wins := 0
	notEligible := 0
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			_, err := db.ClaimTicket(id, "run-"+string(rune('A'+i)), "agent", 90*time.Second)
			mu.Lock()
			defer mu.Unlock()
			switch {
			case err == nil:
				wins++
			case errors.Is(err, ErrNotEligible) || errors.Is(err, ErrAlreadyLeased):
				notEligible++
			default:
				t.Errorf("unexpected claim error: %v", err)
			}
		}(i)
	}
	wg.Wait()

	if wins != 1 {
		t.Fatalf("exactly one claim should win, got %d", wins)
	}
	if wins+notEligible != n {
		t.Fatalf("accounted %d of %d claimants", wins+notEligible, n)
	}

	got, _ := db.GetTicket(id)
	if got.Status != ticket.StatusInProgress {
		t.Errorf("claimed node status = %q, want in_progress", got.Status)
	}
}

func TestClaimRequiresEligibility(t *testing.T) {
	db := openTestDB(t)
	tk := &ticket.Ticket{Title: "node", Status: ticket.StatusBacklog}
	if err := db.CreateTicket(tk, "human"); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ClaimTicket(tk.ID, "run-1", "agent", time.Minute); !errors.Is(err, ErrNotEligible) {
		t.Fatalf("want ErrNotEligible, got %v", err)
	}
}

func TestExpiredLeaseReclaimable(t *testing.T) {
	db := openTestDB(t)
	id := todoNode(t, db)

	// Simulate a stale lease row on a still-eligible node (e.g. a prior run whose
	// lease expired but whose status was reset by recovery). Reclaim must succeed.
	past := encoding.FormatTime(time.Now().Add(-time.Hour))
	if _, err := db.Exec(
		`INSERT INTO leases (ticket_id, run_id, actor_id, status, expires_at, renewed_at)
		 VALUES (?,?,?,?,?,?)`,
		id, "run-old", "agent", leaseActiveStatus, past, past,
	); err != nil {
		t.Fatal(err)
	}

	if _, err := db.ClaimTicket(id, "run-2", "agent", time.Minute); err != nil {
		t.Fatalf("expired lease should be reclaimable: %v", err)
	}
	l, _ := db.GetLease(id)
	if l == nil || l.RunID != "run-2" {
		t.Fatalf("lease not taken over by run-2: %+v", l)
	}
}

func TestRenewAndRelease(t *testing.T) {
	db := openTestDB(t)
	id := todoNode(t, db)
	first, err := db.ClaimTicket(id, "run-1", "agent", time.Minute)
	if err != nil {
		t.Fatal(err)
	}

	renewed, err := db.RenewLease(id, "run-1", 2*time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	if renewed.ExpiresAt <= first.ExpiresAt {
		t.Errorf("renew did not extend expiry: %s !> %s", renewed.ExpiresAt, first.ExpiresAt)
	}

	// Wrong run cannot renew or release.
	if _, err := db.RenewLease(id, "run-2", time.Minute); !errors.Is(err, ErrLeaseNotHeld) {
		t.Errorf("want ErrLeaseNotHeld renewing as wrong run, got %v", err)
	}
	if err := db.ReleaseLease(id, "run-2"); !errors.Is(err, ErrLeaseNotHeld) {
		t.Errorf("want ErrLeaseNotHeld releasing as wrong run, got %v", err)
	}

	if err := db.ReleaseLease(id, "run-1"); err != nil {
		t.Fatal(err)
	}
	l, _ := db.GetLease(id)
	if l != nil {
		t.Errorf("lease should be gone after release: %+v", l)
	}
}
