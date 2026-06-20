package runtime

import (
	"context"
	"testing"
)

func TestStubEmitsLifecycle(t *testing.T) {
	var got []string
	res, err := Stub{}.Run(context.Background(), Spec{RunID: "R-0001", TicketID: "T-0001"}, func(e Event) {
		got = append(got, e.Type)
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if res.Status != "produced" {
		t.Errorf("status = %q, want produced", res.Status)
	}
	want := []string{"claimed", "working", "produced", "awaiting_gate"}
	if len(got) != len(want) {
		t.Fatalf("events = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("event[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestStubHonorsCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := Stub{}.Run(ctx, Spec{TicketID: "T-0001"}, func(Event) {})
	if err == nil {
		t.Fatal("expected cancellation error")
	}
}
