package cli

import "testing"

// gw envelope propose requires a node id and at least one --action, validated
// before any coordinator call (ADR 0054).
func TestEnvelopeProposeRequiresActions(t *testing.T) {
	ctx, _, _ := newTestCtx()
	ctx.RootFlag = projectWithClosedCoordinator(t)

	// Missing --action: rejected as invalid_args (no coordinator needed).
	err := runEnvelopePropose(ctx, []string{"T-1"})
	var ce *Error
	if !asError(err, &ce) || ce.Code != "invalid_args" {
		t.Fatalf("missing --action err = %v, want invalid_args", err)
	}

	// With an action but no reachable coordinator: surfaces coordinator_required,
	// confirming propose is coordinator-mediated (not a direct store write).
	err = runEnvelopePropose(ctx, []string{"T-1", "--action", "execute_children"})
	if !asError(err, &ce) || ce.Code != "coordinator_required" {
		t.Fatalf("with action err = %v, want coordinator_required", err)
	}
}
