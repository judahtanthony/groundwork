package ticket

import "testing"

func leaf(s Status) Rollup { return ComputeRollup(s, nil) }

func TestRollupAllDone(t *testing.T) {
	r := ComputeRollup(StatusInProgress, []Rollup{leaf(StatusDone), leaf(StatusDone)})
	if r.Status != StatusDone {
		t.Errorf("status = %q, want done", r.Status)
	}
}

func TestRollupBlockedSurfaces(t *testing.T) {
	r := ComputeRollup(StatusTodo, []Rollup{leaf(StatusDone), leaf(StatusBlocked), leaf(StatusInProgress)})
	if r.Status != StatusBlocked {
		t.Errorf("status = %q, want blocked (precedence over in_progress)", r.Status)
	}
	if !r.HasBlocked || !r.HasActive {
		t.Errorf("both blocked and active children should be visible: %+v", r)
	}
}

func TestRollupActive(t *testing.T) {
	r := ComputeRollup(StatusTodo, []Rollup{leaf(StatusDone), leaf(StatusInProgress)})
	if r.Status != StatusInProgress {
		t.Errorf("status = %q, want in_progress", r.Status)
	}
}

func TestRollupTodoWhenUnfinishedIdle(t *testing.T) {
	r := ComputeRollup(StatusBacklog, []Rollup{leaf(StatusDone), leaf(StatusTodo)})
	if r.Status != StatusTodo {
		t.Errorf("status = %q, want todo", r.Status)
	}
}

func TestRollupNestedBlockedPropagates(t *testing.T) {
	// grandparent -> parent -> {done, blocked}
	inner := ComputeRollup(StatusTodo, []Rollup{leaf(StatusDone), leaf(StatusBlocked)})
	outer := ComputeRollup(StatusTodo, []Rollup{inner, leaf(StatusDone)})
	if !outer.HasBlocked {
		t.Errorf("blocked descendant should propagate up: %+v", outer)
	}
}

func TestRollupLeafIsOwnStatus(t *testing.T) {
	if leaf(StatusTodo).Status != StatusTodo {
		t.Error("leaf rollup should be its own status")
	}
}

func TestRollupAllCancelledIsCancelled(t *testing.T) {
	r := ComputeRollup(StatusInProgress, []Rollup{leaf(StatusCancelled), leaf(StatusCancelled)})
	if r.Status != StatusCancelled {
		t.Errorf("status = %q, want cancelled (not done) when all children cancelled", r.Status)
	}
}

func TestRollupMixedDoneAndCancelledIsDone(t *testing.T) {
	r := ComputeRollup(StatusInProgress, []Rollup{leaf(StatusDone), leaf(StatusCancelled)})
	if r.Status != StatusDone {
		t.Errorf("status = %q, want done when children are done+cancelled", r.Status)
	}
}
