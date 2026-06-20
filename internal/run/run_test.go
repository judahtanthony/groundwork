package run

import "testing"

func TestStatusValidAndTerminal(t *testing.T) {
	if !StatusRunning.Valid() || Status("bogus").Valid() {
		t.Error("Valid wrong")
	}
	if !StatusCompleted.Terminal() || !StatusCancelled.Terminal() || StatusRunning.Terminal() {
		t.Error("Terminal wrong")
	}
}

func TestModeValid(t *testing.T) {
	if !ModePlanning.Valid() || !ModeImplementation.Valid() || Mode("x").Valid() {
		t.Error("Mode.Valid wrong")
	}
}

func TestCanTransition(t *testing.T) {
	ok := [][2]Status{
		{StatusPending, StatusRunning},
		{StatusRunning, StatusPaused},
		{StatusPaused, StatusRunning},
		{StatusRunning, StatusCompleted},
		{StatusRunning, StatusInterrupted},
		{StatusInterrupted, StatusRunning},
		{StatusRunning, StatusRunning}, // no-op
	}
	for _, p := range ok {
		if !CanTransition(p[0], p[1]) {
			t.Errorf("expected %s->%s legal", p[0], p[1])
		}
	}
	bad := [][2]Status{
		{StatusCompleted, StatusRunning},
		{StatusCancelled, StatusRunning},
		{StatusPending, StatusCompleted},
		{StatusPaused, StatusCompleted},
	}
	for _, p := range bad {
		if CanTransition(p[0], p[1]) {
			t.Errorf("expected %s->%s illegal", p[0], p[1])
		}
	}
}
