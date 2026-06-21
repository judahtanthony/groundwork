package scheduler

import (
	"testing"

	"groundwork/internal/runtime"
	"groundwork/internal/ticket"
)

func sign(n int) int {
	switch {
	case n < 0:
		return -1
	case n > 0:
		return 1
	default:
		return 0
	}
}

func TestComparePath(t *testing.T) {
	p := func(items ...levelKey) []levelKey { return items }
	cases := []struct {
		name string
		a, b []levelKey
		want int
	}{
		{"higher priority first", p(levelKey{0.8, "T-2"}), p(levelKey{0.2, "T-1"}), -1},
		{"equal priority: lower id first (DFS/FIFO)", p(levelKey{0.5, "T-1"}), p(levelKey{0.5, "T-2"}), -1},
		{"priority dominates id", p(levelKey{0.2, "T-9"}), p(levelKey{0.8, "T-1"}), 1},
		{"diverge at root: higher-priority initiative first",
			p(levelKey{0.9, "R1"}, levelKey{0, "X"}), p(levelKey{0.1, "R2"}, levelKey{0, "Y"}), -1},
		// The case that distinguishes per-level interleaving from a naive
		// "all-priorities-then-all-ids" or length tiebreak: equal-priority roots
		// order by root id, not by which path is shorter.
		{"equal-priority roots: lower root id first",
			p(levelKey{0.5, "A"}, levelKey{0, "P"}), p(levelKey{0.5, "B"}), -1},
		{"prefix: ancestor before descendant when equal",
			p(levelKey{0.5, "A"}), p(levelKey{0.5, "A"}, levelKey{0, "P"}), -1},
	}
	for _, c := range cases {
		if got := sign(comparePath(c.a, c.b)); got != c.want {
			t.Errorf("%s: sign(comparePath)=%d, want %d", c.name, got, c.want)
		}
		// Antisymmetry.
		if got := sign(comparePath(c.b, c.a)); got != -c.want {
			t.Errorf("%s: reversed sign=%d, want %d", c.name, got, -c.want)
		}
	}
}

func TestOrderByValueOrdersEligible(t *testing.T) {
	db := newDB(t)
	s := New(db, allowCodexPolicy(), testRegistry(), runtime.Stub{}, nil, testConfig())

	mk := func(parent string, prio float64) *ticket.Ticket {
		tk := &ticket.Ticket{Title: "n", Status: ticket.StatusTodo, ParentID: parent, WorkType: "technical_implementation"}
		if prio > 0 {
			p := prio
			tk.Priority = &p
		}
		if err := db.CreateTicket(tk, "t"); err != nil {
			t.Fatalf("create: %v", err)
		}
		return tk
	}

	root := mk("", 0)
	lo := mk(root.ID, 0.2)
	hi := mk(root.ID, 0.8)

	eligible, err := db.ListEligible()
	if err != nil {
		t.Fatal(err)
	}
	s.orderByValue(eligible)

	pos := map[string]int{}
	for i, n := range eligible {
		pos[n.ID] = i
	}
	if pos[hi.ID] > pos[lo.ID] {
		t.Errorf("high-priority node ran after low: hi@%d lo@%d", pos[hi.ID], pos[lo.ID])
	}
}
