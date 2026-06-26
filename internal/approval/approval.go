// Package approval holds the approval domain (gate decisions that need a
// recorded outcome) and the coordinator service that creates approvals from the
// gate engine and records decisions (trust-and-approvals.md, ADR 0028). An
// approval unlocks a specific gated action and is auditable.
package approval

// Status is an approval's decision state.
type Status string

const (
	StatusPending    Status = "pending"
	StatusApproved   Status = "approved"
	StatusRejected   Status = "rejected"
	StatusClarifying Status = "clarifying" // agent asked to provide more detail
)

// Valid reports whether s is a recognized status.
func (s Status) Valid() bool {
	switch s {
	case StatusPending, StatusApproved, StatusRejected, StatusClarifying:
		return true
	}
	return false
}

// Terminal reports whether s is a final decision.
func (s Status) Terminal() bool {
	return s == StatusApproved || s == StatusRejected
}

// Type is the gated action an approval covers.
type Type string

const (
	TypeExecute    Type = "execute"
	TypeLandToMain Type = "land_to_main"
	TypeDecompose  Type = "decompose"
	TypeReplan     Type = "replan"
	// Authority-elevation as first-class gated actions (ADR 0038): amending policy
	// or raising an autonomy level are ordinary gated actions, default
	// require_human, rather than human-only carve-outs in code. This makes
	// delegating the "improvement" layer expressible without enabling it.
	TypeAmendPolicy     Type = "amend_policy"
	TypeElevateAutonomy Type = "elevate_autonomy"
)

// Valid reports whether t is a recognized approval type.
func (t Type) Valid() bool {
	switch t {
	case TypeExecute, TypeLandToMain, TypeDecompose, TypeReplan,
		TypeAmendPolicy, TypeElevateAutonomy:
		return true
	}
	return false
}
