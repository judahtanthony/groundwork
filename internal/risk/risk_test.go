package risk

import "testing"

func TestReversibleEmptyScope(t *testing.T) {
	ok, reasons := Reversible(Scope{Files: []string{"docs/x.md"}})
	if !ok {
		t.Errorf("empty scope reversible=false, reasons=%v", reasons)
	}
}

func TestReversibleTriggers(t *testing.T) {
	cases := []struct {
		name  string
		scope Scope
	}{
		{"external", Scope{External: true}},
		{"migration", Scope{IrreversibleMigration: true}},
		{"credential_flag", Scope{CredentialAccess: true}},
		{"secret_file", Scope{Files: []string{"config/.env.production"}}},
		{"secret_named", Scope{Files: []string{"deploy/db_secret.txt"}}},
		{"destructive", Scope{Commands: []string{"rm -rf /tmp/x"}}},
		{"force_push", Scope{Commands: []string{"git push origin main --force"}}},
		{"drop_table", Scope{Commands: []string{"psql -c 'DROP TABLE users'"}}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ok, reasons := Reversible(tc.scope)
			if ok || len(reasons) == 0 {
				t.Errorf("expected irreversible with reasons, got ok=%v reasons=%v", ok, reasons)
			}
		})
	}
}

func TestClassForScoreBoundaries(t *testing.T) {
	cases := []struct {
		score int
		want  Class
	}{
		{0, ClassLow}, {33, ClassLow}, {34, ClassMedium}, {66, ClassMedium},
		{67, ClassHigh}, {100, ClassHigh},
	}
	for _, tc := range cases {
		if got := ClassForScore(tc.score); got != tc.want {
			t.Errorf("ClassForScore(%d) = %s, want %s", tc.score, got, tc.want)
		}
	}
}

func TestScoreNeverExceeds100(t *testing.T) {
	s := Scope{
		Files:                 []string{"a", "b", "c", "d", "e", "f", "g", "h"},
		Commands:              []string{"rm -rf /"},
		Network:               true,
		External:              true,
		IrreversibleMigration: true,
		CredentialAccess:      true,
	}
	if got := Score(s); got != 100 {
		t.Errorf("Score = %d, want capped at 100", got)
	}
}

func TestScoreClassesEscalate(t *testing.T) {
	low := ClassForScore(Score(Scope{Files: []string{"docs/a.md"}}))
	if low != ClassLow {
		t.Errorf("docs-only class = %s, want low", low)
	}
	high := ClassForScore(Score(Scope{External: true}))
	if high == ClassLow {
		t.Errorf("external-state class = %s, want > low", high)
	}
}

func TestClassAtMost(t *testing.T) {
	if !ClassLow.AtMost(ClassMedium) {
		t.Error("low should be at most medium")
	}
	if ClassHigh.AtMost(ClassMedium) {
		t.Error("high should not be at most medium")
	}
	if !ClassMedium.AtMost(ClassMedium) {
		t.Error("medium should be at most medium")
	}
}
