package sqlite

import (
	"testing"

	"groundwork/internal/ticket"
)

// A work type with enough clean done leaves yields one pending elevation
// suggestion; rescanning is idempotent and dismissal clears it (ADR 0038).
func TestGenerateElevationSuggestions(t *testing.T) {
	db := openTestDB(t)
	for i := 0; i < SuggestionElevationThreshold; i++ {
		tk := &ticket.Ticket{Title: "doc", NodeType: ticket.NodeLeaf, Status: ticket.StatusDone, WorkType: "documentation"}
		if err := db.CreateTicket(tk, "tester"); err != nil {
			t.Fatal(err)
		}
		if _, err := db.RecordValidation(ValidationResult{TicketID: tk.ID, Name: "go", Status: ValidationPass}); err != nil {
			t.Fatal(err)
		}
	}

	created, err := db.GenerateElevationSuggestions()
	if err != nil {
		t.Fatal(err)
	}
	if len(created) != 1 || created[0].WorkType != "documentation" || created[0].Level != "auto" {
		t.Fatalf("created = %+v, want one documentation->auto elevation", created)
	}

	again, err := db.GenerateElevationSuggestions()
	if err != nil {
		t.Fatal(err)
	}
	if len(again) != 0 {
		t.Errorf("second scan created %d, want 0 (idempotent)", len(again))
	}

	pending, err := db.ListSuggestions("pending")
	if err != nil {
		t.Fatal(err)
	}
	if len(pending) != 1 {
		t.Fatalf("pending = %d, want 1", len(pending))
	}
	if _, err := db.SetSuggestionStatus(pending[0].ID, "dismissed"); err != nil {
		t.Fatal(err)
	}
	if p, _ := db.ListSuggestions("pending"); len(p) != 0 {
		t.Errorf("pending after dismiss = %d, want 0", len(p))
	}
}

// A single failing validation in the cohort blocks the elevation suggestion.
func TestGenerateElevationSuggestionsBlockedByFailure(t *testing.T) {
	db := openTestDB(t)
	for i := 0; i < SuggestionElevationThreshold; i++ {
		tk := &ticket.Ticket{Title: "impl", NodeType: ticket.NodeLeaf, Status: ticket.StatusDone, WorkType: "technical_implementation"}
		if err := db.CreateTicket(tk, "tester"); err != nil {
			t.Fatal(err)
		}
		status := ValidationPass
		if i == SuggestionElevationThreshold-1 {
			status = ValidationFail
		}
		if _, err := db.RecordValidation(ValidationResult{TicketID: tk.ID, Name: "go", Status: status}); err != nil {
			t.Fatal(err)
		}
	}
	created, err := db.GenerateElevationSuggestions()
	if err != nil {
		t.Fatal(err)
	}
	if len(created) != 0 {
		t.Errorf("created %d, want 0 (a failure blocks elevation)", len(created))
	}
}
