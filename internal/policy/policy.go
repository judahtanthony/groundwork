// Package policy loads the trust, autonomy, and validation policies
// (docs/contracts/policies.md) and evaluates the gate decisions that compose
// them with risk and reversibility (ADR 0028). Match conditions are uniform: a
// rule carries one `when:` predicate (the heterogeneous flat/nested shapes in
// the original contract examples are normalized to `when:` per ADR 0028). The
// engine produces decisions, never side effects (overview.md).
package policy

import (
	"fmt"

	"gopkg.in/yaml.v3"
	"groundwork/internal/risk"
)

// Policy schema identifiers.
const (
	TrustSchema      = "groundwork_trust_policy/v1"
	AutonomySchema   = "groundwork_autonomy_policy/v1"
	ValidationSchema = "groundwork_validation_policy/v1"
)

// Match is the parameterized rule predicate (ADR 0028, ADR 0029). Every present
// field must hold for the rule to match; absent fields are wildcards. Identity
// (actor_ids, prefix-matched) and capability/role are independent dimensions of
// the same AND.
type Match struct {
	ActorIDs           []string `yaml:"actor_ids,omitempty"`
	ActorTypes         []string `yaml:"actor_types,omitempty"`
	Roles              []string `yaml:"roles,omitempty"`
	WorkTypes          []string `yaml:"work_types,omitempty"`
	ActionTypes        []string `yaml:"action_types,omitempty"`
	Files              []string `yaml:"files,omitempty"`
	RiskClass          string   `yaml:"risk_class,omitempty"`
	RiskClassAtMost    string   `yaml:"risk_class_at_most,omitempty"`
	Reversible         *bool    `yaml:"reversible,omitempty"`
	CommandRegex       string   `yaml:"command_regex,omitempty"`
	CommandCategories  []string `yaml:"command_categories,omitempty"`
	ChangeType         string   `yaml:"change_type,omitempty"`
	MaxDiffLines       int      `yaml:"max_diff_lines,omitempty"`
	CwdWithinWorkspace *bool    `yaml:"cwd_within_workspace,omitempty"`
	Network            *bool    `yaml:"network,omitempty"`
	// WithinEnvelope matches the coordinator-computed fact that an action sits
	// inside an approved parent/root envelope (ADR 0056), so allow_claim rules can
	// require bounded-autonomy authorization.
	WithinEnvelope *bool `yaml:"within_envelope,omitempty"`
}

// ReviewAllowedBy names who may review/decide for an auto_approve rule that
// delegates to a reviewer (e.g. an ai_judge). Reviewer-agent execution is a
// later phase; the field is parsed now for forward compatibility.
type ReviewAllowedBy struct {
	ActorIDs []string `yaml:"actor_ids,omitempty"`
	Roles    []string `yaml:"roles,omitempty"`
}

// Rule is one trust-policy rule with a stable id (surfaced as R-01 etc.).
type Rule struct {
	ID              string           `yaml:"id"`
	Description     string           `yaml:"description,omitempty"`
	When            Match            `yaml:"when"`
	Actions         []string         `yaml:"actions,omitempty"`
	ReviewAllowedBy *ReviewAllowedBy `yaml:"review_allowed_by,omitempty"`
	// RequireRoles names the approver role(s) a matching require_human rule
	// demands (ADR 0055/0048): the approval then records which role was required
	// and why. In v1 the owner satisfies every role; this keeps the record honest
	// for later multi-human identity.
	RequireRoles []string `yaml:"require_roles,omitempty"`
}

// TrustPolicy is the ordered, first-match trust rule set.
type TrustPolicy struct {
	Schema       string `yaml:"schema"`
	AutoApprove  []Rule `yaml:"auto_approve"`
	RequireHuman []Rule `yaml:"require_human"`
	AllowClaim   []Rule `yaml:"allow_claim"`
}

// AutonomyRequires lists what an elevated autonomy level depends on.
type AutonomyRequires struct {
	SOP         string   `yaml:"sop,omitempty"`
	Validations []string `yaml:"validations,omitempty"`
}

// AutonomyByWorkType is a per-work-type autonomy override.
type AutonomyByWorkType struct {
	Level    string           `yaml:"level"`
	Requires AutonomyRequires `yaml:"requires,omitempty"`
}

// AutonomyAction is the autonomy configuration for one gated action.
type AutonomyAction struct {
	Default    string                        `yaml:"default"`
	ByWorkType map[string]AutonomyByWorkType `yaml:"by_work_type,omitempty"`
}

// AutonomyPolicy holds per-action autonomy levels.
type AutonomyPolicy struct {
	Schema  string                    `yaml:"schema"`
	Actions map[string]AutonomyAction `yaml:"actions"`
}

// ValidationCheck is one required validation command.
type ValidationCheck struct {
	Name    string `yaml:"name"`
	Command string `yaml:"command,omitempty"`
}

// ValidationMatch selects files a template applies to.
type ValidationMatch struct {
	Files []string `yaml:"files,omitempty"`
}

// ValidationTemplate maps changed files to required checks.
type ValidationTemplate struct {
	Match            ValidationMatch   `yaml:"match"`
	Required         []ValidationCheck `yaml:"required"`
	LandingRiskFloor string            `yaml:"landing_risk_floor,omitempty"`
}

// ValidationPolicy holds validation templates by name.
type ValidationPolicy struct {
	Schema    string                        `yaml:"schema"`
	Templates map[string]ValidationTemplate `yaml:"templates"`
}

// validAutonomyLevels are the accepted autonomy levels (reviewer is parsed but
// behaves as human-required until reviewer agents land).
var validAutonomyLevels = map[string]bool{"require_human": true, "reviewer": true, "auto": true}

// ParseTrust decodes and validates a trust policy.
func ParseTrust(data []byte) (*TrustPolicy, []string, error) {
	var p TrustPolicy
	if err := yaml.Unmarshal(data, &p); err != nil {
		return nil, nil, fmt.Errorf("parsing trust policy: %w", err)
	}
	warnings := schemaWarnings(p.Schema, TrustSchema, data, "schema", "auto_approve", "require_human", "allow_claim")

	seen := map[string]bool{}
	for _, group := range [][]Rule{p.AutoApprove, p.RequireHuman, p.AllowClaim} {
		for i := range group {
			if err := validateRule(&group[i], seen); err != nil {
				return nil, warnings, err
			}
		}
	}
	return &p, warnings, nil
}

// validateRule enforces non-empty unique ids and valid risk-class fields.
func validateRule(r *Rule, seen map[string]bool) error {
	if r.ID == "" {
		return fmt.Errorf("trust rule has an empty id")
	}
	if seen[r.ID] {
		return fmt.Errorf("duplicate trust rule id %q", r.ID)
	}
	seen[r.ID] = true
	for _, rc := range []string{r.When.RiskClass, r.When.RiskClassAtMost} {
		if rc != "" && !risk.Class(rc).Valid() {
			return fmt.Errorf("trust rule %q has invalid risk class %q", r.ID, rc)
		}
	}
	return nil
}

// ParseAutonomy decodes and validates an autonomy policy.
func ParseAutonomy(data []byte) (*AutonomyPolicy, []string, error) {
	var p AutonomyPolicy
	if err := yaml.Unmarshal(data, &p); err != nil {
		return nil, nil, fmt.Errorf("parsing autonomy policy: %w", err)
	}
	warnings := schemaWarnings(p.Schema, AutonomySchema, data, "schema", "actions")
	for name, a := range p.Actions {
		if a.Default != "" && !validAutonomyLevels[a.Default] {
			return nil, warnings, fmt.Errorf("autonomy action %q has invalid default level %q", name, a.Default)
		}
		for wt, by := range a.ByWorkType {
			if !validAutonomyLevels[by.Level] {
				return nil, warnings, fmt.Errorf("autonomy action %q work type %q has invalid level %q", name, wt, by.Level)
			}
		}
	}
	return &p, warnings, nil
}

// ParseValidation decodes and validates a validation policy.
func ParseValidation(data []byte) (*ValidationPolicy, []string, error) {
	var p ValidationPolicy
	if err := yaml.Unmarshal(data, &p); err != nil {
		return nil, nil, fmt.Errorf("parsing validation policy: %w", err)
	}
	warnings := schemaWarnings(p.Schema, ValidationSchema, data, "schema", "templates")
	for name, tmpl := range p.Templates {
		if tmpl.LandingRiskFloor != "" && !risk.Class(tmpl.LandingRiskFloor).Valid() {
			return nil, warnings, fmt.Errorf("validation template %q has invalid landing_risk_floor %q", name, tmpl.LandingRiskFloor)
		}
		for i, c := range tmpl.Required {
			if c.Name == "" {
				return nil, warnings, fmt.Errorf("validation template %q check #%d has an empty name", name, i+1)
			}
		}
	}
	return &p, warnings, nil
}

// schemaWarnings reports a schema mismatch and any unknown top-level keys,
// mirroring the forward-compatible warn-don't-fail policy of config/actor.
func schemaWarnings(got, want string, data []byte, known ...string) []string {
	var warnings []string
	if got != want {
		warnings = append(warnings, fmt.Sprintf("policy schema %q does not match expected %q", got, want))
	}
	knownSet := map[string]bool{}
	for _, k := range known {
		knownSet[k] = true
	}
	var top map[string]yaml.Node
	if err := yaml.Unmarshal(data, &top); err == nil {
		for k := range top {
			if !knownSet[k] {
				warnings = append(warnings, fmt.Sprintf("unknown policy key %q (ignored)", k))
			}
		}
	}
	return warnings
}
