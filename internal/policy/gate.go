package policy

import (
	"regexp"
	"strings"

	"groundwork/internal/actor"
	"groundwork/internal/risk"
)

// Action is the context a gate decision is evaluated against.
type Action struct {
	Type               string       // execute | land_to_main | decompose | replan | review | claim
	Actor              *actor.Actor // resolved instance; nil if unresolved
	WorkType           string
	ChangeType         string // e.g. documentation; "" when unknown
	DiffLines          int    // 0 when unknown
	CwdWithinWorkspace bool
	Scope              risk.Scope
}

// Outcome is a gate decision result.
type Outcome string

const (
	OutcomeAutoApprove  Outcome = "auto_approve"
	OutcomeRequireHuman Outcome = "require_human"
	OutcomeAllow        Outcome = "allow" // claim/route authorization granted
	OutcomeDeny         Outcome = "deny"
)

// Decision is the explainable result of a gate evaluation (ADR 0028): the
// outcome, the fired rule id (empty when none/autonomy default), and the risk
// and reversibility verdicts shown on the approval surface.
type Decision struct {
	Outcome    Outcome
	RuleID     string
	RiskClass  risk.Class
	RiskScore  int
	Reversible bool
	Reasons    []string
	// RequiredRoles is the approver role(s) the firing require_human rule demands
	// (ADR 0055). Empty when no rule named one; the approval records it so the
	// gate decision is explainable and ready for multi-human identity.
	RequiredRoles []string
}

// classify computes the risk score, class, and reversibility for an action. An
// irreversible action is class critical regardless of score (ADR 0014).
func classify(a Action) (risk.Class, int, bool, []string) {
	score := risk.Score(a.Scope)
	reversible, reasons := risk.Reversible(a.Scope)
	class := risk.ClassForScore(score)
	if !reversible {
		class = risk.ClassCritical
	}
	return class, score, reversible, reasons
}

// Evaluate decides whether a gated action (execute, land_to_main, decompose, …)
// may proceed automatically or needs a human. Composition order (ADR 0028):
// reversibility floor → first-match require_human → first-match auto_approve →
// autonomy default. Earlier steps cannot be loosened by later ones.
func (s *Set) Evaluate(a Action) Decision {
	class, score, reversible, reasons := classify(a)
	d := Decision{RiskClass: class, RiskScore: score, Reversible: reversible, Reasons: reasons}

	// 1. Reversibility floor: irreversible is forced critical, human-required.
	if !reversible {
		d.Outcome = OutcomeRequireHuman
		return d
	}

	if s.Trust != nil {
		// 2. require_human wins over any auto path.
		if r := firstMatch(s.Trust.RequireHuman, a, class, reversible); r != nil {
			d.Outcome = OutcomeRequireHuman
			d.RuleID = r.ID
			d.RequiredRoles = r.RequireRoles
			return d
		}
		// 3. auto_approve.
		if r := firstMatch(s.Trust.AutoApprove, a, class, reversible); r != nil {
			d.Outcome = OutcomeAutoApprove
			d.RuleID = r.ID
			return d
		}
	}

	// 4. Autonomy default for this action (with per-work-type elevation).
	d.Outcome = s.autonomyOutcome(a)
	return d
}

// AuthorizeClaim decides whether the actor may claim/act on the node. It scans
// allow_claim top-down; the first rule that matches and whose actions include
// the action type (or lists no actions) grants the claim. No match is a deny
// (ADR 0028 default-deny). Reversibility/risk are reported for explanation but
// do not themselves authorize a claim.
func (s *Set) AuthorizeClaim(a Action) Decision {
	class, score, reversible, reasons := classify(a)
	d := Decision{Outcome: OutcomeDeny, RiskClass: class, RiskScore: score, Reversible: reversible, Reasons: reasons}
	if s.Trust == nil {
		return d
	}
	for i := range s.Trust.AllowClaim {
		r := &s.Trust.AllowClaim[i]
		if matches(&r.When, a, class, reversible) && actionAllowed(r.Actions, a.Type) {
			d.Outcome = OutcomeAllow
			d.RuleID = r.ID
			return d
		}
	}
	return d
}

// autonomyOutcome resolves the autonomy level for the action (per-work-type
// override, else the action default, else require_human) and maps it to an
// outcome. reviewer behaves as human-required until reviewer agents land.
func (s *Set) autonomyOutcome(a Action) Outcome {
	level := "require_human"
	if s.Autonomy != nil {
		if act, ok := s.Autonomy.Actions[a.Type]; ok {
			if act.Default != "" {
				level = act.Default
			}
			if by, ok := act.ByWorkType[a.WorkType]; ok && by.Level != "" {
				level = by.Level
			}
		}
	}
	if level == "auto" {
		return OutcomeAutoApprove
	}
	return OutcomeRequireHuman
}

// firstMatch returns the first rule whose predicate matches, or nil.
func firstMatch(rules []Rule, a Action, class risk.Class, reversible bool) *Rule {
	for i := range rules {
		if matches(&rules[i].When, a, class, reversible) {
			return &rules[i]
		}
	}
	return nil
}

// actionAllowed reports whether actionType is permitted by a rule's actions
// list. An empty list means the rule applies to any action.
func actionAllowed(actions []string, actionType string) bool {
	if len(actions) == 0 {
		return true
	}
	for _, x := range actions {
		if x == actionType {
			return true
		}
	}
	return false
}

// matches reports whether every present condition in m holds for the action.
// Absent conditions are wildcards. Fields requiring change metadata not yet
// available in v1 (change_type, max_diff_lines, cwd_within_workspace) are
// evaluated only when the action supplies them, and skipped otherwise.
func matches(m *Match, a Action, class risk.Class, reversible bool) bool {
	if len(m.ActorIDs) > 0 {
		if a.Actor == nil || !actor.AnyIDMatch(m.ActorIDs, a.Actor.ID) {
			return false
		}
	}
	if len(m.ActorTypes) > 0 {
		if a.Actor == nil || !inList(m.ActorTypes, string(a.Actor.Type)) {
			return false
		}
	}
	if len(m.Roles) > 0 {
		if a.Actor == nil || !a.Actor.HasAnyRole(m.Roles) {
			return false
		}
	}
	if len(m.WorkTypes) > 0 && !inList(m.WorkTypes, a.WorkType) {
		return false
	}
	if len(m.ActionTypes) > 0 && !inList(m.ActionTypes, a.Type) {
		return false
	}
	if len(m.Files) > 0 && !anyFileMatch(m.Files, a.Scope.Files) {
		return false
	}
	if m.RiskClass != "" && string(class) != m.RiskClass {
		return false
	}
	if m.RiskClassAtMost != "" && !class.AtMost(risk.Class(m.RiskClassAtMost)) {
		return false
	}
	if m.Reversible != nil && reversible != *m.Reversible {
		return false
	}
	if m.CommandRegex != "" && !anyCommandMatch(m.CommandRegex, a.Scope.Commands) {
		return false
	}
	if len(m.CommandCategories) > 0 && !commandCategoryMatch(m.CommandCategories, a.Scope) {
		return false
	}
	if m.Network != nil && a.Scope.Network != *m.Network {
		return false
	}
	if m.ChangeType != "" && a.ChangeType != m.ChangeType {
		return false
	}
	if m.MaxDiffLines > 0 && a.DiffLines > m.MaxDiffLines {
		return false
	}
	if m.CwdWithinWorkspace != nil && a.CwdWithinWorkspace != *m.CwdWithinWorkspace {
		return false
	}
	return true
}

func inList(list []string, v string) bool {
	for _, x := range list {
		if x == v {
			return true
		}
	}
	return false
}

func anyCommandMatch(pattern string, cmds []string) bool {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return false
	}
	for _, c := range cmds {
		if re.MatchString(c) {
			return true
		}
	}
	return false
}

// commandCategoryMatch supports the "destructive" category in v1 (the others are
// forward-compatible no-ops).
func commandCategoryMatch(categories []string, scope risk.Scope) bool {
	for _, c := range categories {
		if c == "destructive" && risk.HasDestructiveCommand(scope) {
			return true
		}
	}
	return false
}

// anyFileMatch reports whether any file matches any glob pattern.
func anyFileMatch(patterns, files []string) bool {
	for _, p := range patterns {
		re := globToRegexp(p)
		for _, f := range files {
			if re.MatchString(f) {
				return true
			}
		}
	}
	return false
}

// globToRegexp converts a path glob to an anchored regexp. `**/` matches zero or
// more leading directories, `**` matches any run including separators, `*`
// matches within a path segment, and `?` matches one non-separator char.
func globToRegexp(glob string) *regexp.Regexp {
	var b strings.Builder
	b.WriteString("^")
	for i := 0; i < len(glob); i++ {
		switch {
		case strings.HasPrefix(glob[i:], "**/"):
			b.WriteString("(?:.*/)?")
			i += 2
		case glob[i] == '*' && i+1 < len(glob) && glob[i+1] == '*':
			b.WriteString(".*")
			i++
		case glob[i] == '*':
			b.WriteString("[^/]*")
		case glob[i] == '?':
			b.WriteString("[^/]")
		default:
			b.WriteString(regexp.QuoteMeta(string(glob[i])))
		}
	}
	b.WriteString("$")
	re, err := regexp.Compile(b.String())
	if err != nil {
		// A malformed glob matches nothing rather than panicking.
		return regexp.MustCompile(`$.^`)
	}
	return re
}
