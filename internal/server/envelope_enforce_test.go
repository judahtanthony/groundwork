package server

import (
	"errors"
	"testing"

	"groundwork/internal/approval"
	"groundwork/internal/envelope"
	"groundwork/internal/store/sqlite"
	"groundwork/internal/ticket"
)

func TestEnvelopeEnforceHelpers(t *testing.T) {
	if !touchesContracts([]string{"src/a.go", "docs/contracts/x.md"}) {
		t.Error("contract path not detected")
	}
	if touchesContracts([]string{"src/a.go"}) {
		t.Error("false contract detection")
	}
	if !touchesPublicAPI([]string{"internal/api/routes.go"}) || !touchesPublicAPI([]string{"schema.proto"}) {
		t.Error("public API path not detected")
	}
	if touchesPublicAPI([]string{"internal/store/db.go"}) {
		t.Error("false public-API detection")
	}
	env := &envelope.Envelope{Scope: envelope.Scope{Files: envelope.FileScope{Allow: []string{"src/**"}}}}
	if !hasUnexpectedFiles(env, []string{"other/x.go"}) {
		t.Error("unexpected file not detected")
	}
	if hasUnexpectedFiles(env, []string{"src/a.go"}) {
		t.Error("in-scope file flagged unexpected")
	}
	// An empty allow-list means anything goes.
	if hasUnexpectedFiles(&envelope.Envelope{}, []string{"anywhere.go"}) {
		t.Error("empty allow-list should not flag files")
	}
}

// setRunDiff records a completed run with a changed-file set for a node.
func setRunDiff(t *testing.T, db *sqlite.DB, nodeID, runID string, files []string) {
	t.Helper()
	if _, err := db.Exec(`INSERT INTO runs
		(id, ticket_id, actor_id, actor_snapshot_json, mode, runtime, model, status, workspace_path, started_at, updated_at)
		VALUES (?,?,?,?,?,?,?,?,?,?,?)`,
		runID, nodeID, "ai.codex.default", "{}", "implementation", "codex", "m", "completed", "", "2026-06-28T10:00:00Z", "2026-06-28T10:00:00Z"); err != nil {
		t.Fatal(err)
	}
	if err := db.SetRunChangedFiles(runID, files); err != nil {
		t.Fatal(err)
	}
}

func activeEnvelope(t *testing.T, db *sqlite.DB, nodeID string, e *envelope.Envelope) {
	t.Helper()
	e.ID = "AE-" + nodeID
	e.NodeID = nodeID
	e.Status = envelope.StatusActive
	if e.ApprovedActions == nil {
		e.ApprovedActions = []string{envelope.ActionLandChildToParent}
	}
	if err := db.UpsertEnvelope(e); err != nil {
		t.Fatal(err)
	}
}

func TestEnforceEnvelopeWithinScopeIsNoOp(t *testing.T) {
	srv, db := newTestServer(t)
	node := &ticket.Ticket{Title: "n", NodeType: ticket.NodeLeaf, Status: ticket.StatusInProgress, WorkType: "technical_implementation"}
	if err := db.CreateTicket(node, "tester"); err != nil {
		t.Fatal(err)
	}
	activeEnvelope(t, db, node.ID, &envelope.Envelope{
		Scope:      envelope.Scope{Files: envelope.FileScope{Allow: []string{"src/**"}}},
		Escalation: envelope.Escalation{OnUnexpectedFiles: true, OnContractChange: true, OnValidationFailure: true},
	})
	setRunDiff(t, db, node.ID, "R-1", []string{"src/feature.go"})

	appr, err := srv.enforceEnvelopeOnDiff(node.ID, "land_to_parent")
	if err != nil || appr != nil {
		t.Fatalf("in-scope clean diff should not escalate: appr=%v err=%v", appr, err)
	}
}

func TestEnforceEnvelopeEscalatesOnTriggers(t *testing.T) {
	cases := []struct {
		name  string
		env   envelope.Escalation
		allow []string
		files []string
		fail  bool // record a failing validation
	}{
		{"unexpected files", envelope.Escalation{OnUnexpectedFiles: true}, []string{"src/**"}, []string{"other/x.go"}, false},
		{"contract change", envelope.Escalation{OnContractChange: true}, []string{"**"}, []string{"docs/contracts/x.md"}, false},
		{"public api change", envelope.Escalation{OnPublicAPIChange: true}, []string{"**"}, []string{"internal/api/routes.go"}, false},
		{"validation failure", envelope.Escalation{OnValidationFailure: true}, []string{"**"}, []string{"src/a.go"}, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			srv, db := newTestServer(t)
			node := &ticket.Ticket{Title: "n", NodeType: ticket.NodeLeaf, Status: ticket.StatusInProgress, WorkType: "technical_implementation"}
			if err := db.CreateTicket(node, "tester"); err != nil {
				t.Fatal(err)
			}
			activeEnvelope(t, db, node.ID, &envelope.Envelope{
				Scope:      envelope.Scope{Files: envelope.FileScope{Allow: tc.allow}},
				Escalation: tc.env,
			})
			setRunDiff(t, db, node.ID, "R-1", tc.files)
			if tc.fail {
				if _, err := db.RecordValidation(sqlite.ValidationResult{TicketID: node.ID, Name: "go test", Status: sqlite.ValidationFail}); err != nil {
					t.Fatal(err)
				}
			}

			appr, err := srv.enforceEnvelopeOnDiff(node.ID, "land_to_parent")
			if !errors.Is(err, ErrEnvelopeEscalation) {
				t.Fatalf("err = %v, want ErrEnvelopeEscalation", err)
			}
			if appr == nil || appr.Type != string(approval.TypeException) {
				t.Fatalf("expected an exception approval, got %+v", appr)
			}
		})
	}
}

func TestEnforceEnvelopeNoEnvelopeIsNoOp(t *testing.T) {
	srv, db := newTestServer(t)
	node := &ticket.Ticket{Title: "n", NodeType: ticket.NodeLeaf, Status: ticket.StatusInProgress, WorkType: "technical_implementation"}
	if err := db.CreateTicket(node, "tester"); err != nil {
		t.Fatal(err)
	}
	setRunDiff(t, db, node.ID, "R-1", []string{"anything.go"})
	appr, err := srv.enforceEnvelopeOnDiff(node.ID, "land_to_parent")
	if err != nil || appr != nil {
		t.Fatalf("ungoverned node should be a no-op: appr=%v err=%v", appr, err)
	}
}
