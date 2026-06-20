package risk

// Class is a gate risk class. Gates key off the class, not the raw score
// (trust-and-approvals.md).
type Class string

const (
	ClassLow      Class = "low"
	ClassMedium   Class = "medium"
	ClassHigh     Class = "high"
	ClassCritical Class = "critical"
)

// Valid reports whether c is a recognized class.
func (c Class) Valid() bool {
	switch c {
	case ClassLow, ClassMedium, ClassHigh, ClassCritical:
		return true
	}
	return false
}

// AtMost reports whether c is no riskier than max, using the class ordering
// low < medium < high < critical. Used by `risk_class_at_most` rule matching.
func (c Class) AtMost(max Class) bool {
	return classRank(c) <= classRank(max)
}

func classRank(c Class) int {
	switch c {
	case ClassLow:
		return 0
	case ClassMedium:
		return 1
	case ClassHigh:
		return 2
	case ClassCritical:
		return 3
	default:
		return 0
	}
}

// Score maps a scope to a 0–100 risk score (display/ranking only). It is a
// coarse additive heuristic; calibrated scoring is a Phase 5 refinement.
func Score(s Scope) int {
	score := min(len(s.Files)*5, 30)
	if s.Network {
		score += 15
	}
	if s.External {
		score += 40
	}
	if s.IrreversibleMigration {
		score += 30
	}
	if len(destructiveCommands(s.Commands)) > 0 {
		score += 40
	}
	if s.CredentialAccess || hasSecretFile(s.Files) {
		score += 30
	}
	return min(score, 100)
}

// ClassForScore maps a 0–100 score onto a named class (trust-and-approvals.md):
// low 0–33, medium 34–66, high 67–100. It never returns critical: critical is
// forced by irreversibility or an explicit rule at the gate (ADR 0028), not by
// the score.
func ClassForScore(score int) Class {
	switch {
	case score <= 33:
		return ClassLow
	case score <= 66:
		return ClassMedium
	default:
		return ClassHigh
	}
}
