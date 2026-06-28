// Package envelope is the approval-envelope domain (ADR 0054): a human-approved
// boundary attached to a composite/root node within which bounded child planning,
// execution, and land-to-parent may proceed. The envelope is file-authoritative —
// it lives as a per-node sidecar (.groundwork/tickets/<id>/envelope.yaml) and is
// mirrored into SQLite for live evaluation (ADR 0040/0053).
package envelope

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Status is an envelope's lifecycle state.
type Status string

const (
	StatusActive     Status = "active"
	StatusRevoked    Status = "revoked"
	StatusSuperseded Status = "superseded"
)

// Approved-action vocabulary an envelope may grant (a subset of the gate action
// space; ADR 0028/0038). Root land_to_main, policy changes, and autonomy
// elevation are never grantable by an envelope.
const (
	ActionDecomposeChildren = "decompose_children"
	ActionExecuteChildren   = "execute_children"
	ActionLandChildToParent = "land_children_to_parent"
	ActionReplanWithinGoal  = "replan_within_goal"
)

// KnownActions is the full approved-action vocabulary an envelope may grant.
var KnownActions = []string{
	ActionDecomposeChildren, ActionExecuteChildren,
	ActionLandChildToParent, ActionReplanWithinGoal,
}

// ValidateActions rejects any approved action outside the known vocabulary, so a
// malformed envelope is refused at activation rather than silently granting (or
// failing to grant) an unrecognized action (ADR 0054/0056). An envelope granting
// no actions is also rejected — it would authorize nothing.
func ValidateActions(actions []string) error {
	if len(actions) == 0 {
		return fmt.Errorf("envelope grants no approved_actions")
	}
	known := map[string]bool{}
	for _, a := range KnownActions {
		known[a] = true
	}
	var unknown []string
	for _, a := range actions {
		if !known[a] {
			unknown = append(unknown, a)
		}
	}
	if len(unknown) > 0 {
		return fmt.Errorf("unknown envelope approved_actions %v (known: %v)", unknown, KnownActions)
	}
	return nil
}

// Planning bounds dynamic decomposition inside the envelope.
type Planning struct {
	MaxDepth         int      `yaml:"max_depth" json:"max_depth"`
	MaxChildren      int      `yaml:"max_children" json:"max_children"`
	AllowedWorkTypes []string `yaml:"allowed_work_types" json:"allowed_work_types"`
}

// FileScope is the v1 (file-glob only) resource scope.
type FileScope struct {
	Allow         []string `yaml:"allow" json:"allow"`
	RequireReview []string `yaml:"require_review" json:"require_review"`
	Deny          []string `yaml:"deny" json:"deny"`
}

// Scope is the envelope's resource scope (files only in v1; ADR 0046 symbols/
// resources later).
type Scope struct {
	Files FileScope `yaml:"files" json:"files"`
}

// Validation lists the validation templates required inside the envelope.
type Validation struct {
	RequiredTemplates []string `yaml:"required_templates" json:"required_templates"`
}

// Escalation lists the boundary-crossing triggers that raise an exception
// approval rather than proceeding (ADR 0054/0056).
type Escalation struct {
	OnUnexpectedFiles   bool `yaml:"on_unexpected_files" json:"on_unexpected_files"`
	OnContractChange    bool `yaml:"on_contract_change" json:"on_contract_change"`
	OnValidationFailure bool `yaml:"on_validation_failure" json:"on_validation_failure"`
	OnRiskAboveCeiling  bool `yaml:"on_risk_above_ceiling" json:"on_risk_above_ceiling"`
	OnPublicAPIChange   bool `yaml:"on_public_api_change" json:"on_public_api_change"`
}

// Envelope is the approved boundary record (ADR 0054 shape).
type Envelope struct {
	ID              string     `yaml:"id" json:"id"`
	NodeID          string     `yaml:"node_id" json:"node_id"`
	Status          Status     `yaml:"status" json:"status"`
	ApprovedBy      string     `yaml:"approved_by" json:"approved_by"`
	ApprovedAt      string     `yaml:"approved_at" json:"approved_at"`
	ApprovedActions []string   `yaml:"approved_actions" json:"approved_actions"`
	Planning        Planning   `yaml:"planning" json:"planning"`
	Scope           Scope      `yaml:"scope" json:"scope"`
	Validation      Validation `yaml:"validation" json:"validation"`
	RiskCeiling     string     `yaml:"risk_ceiling" json:"risk_ceiling"`
	AllowedRoles    []string   `yaml:"allowed_roles" json:"allowed_roles"`
	Escalation      Escalation `yaml:"escalation" json:"escalation"`
}

// Allows reports whether the envelope grants the given approved action.
func (e *Envelope) Allows(action string) bool {
	for _, a := range e.ApprovedActions {
		if a == action {
			return true
		}
	}
	return false
}

// AllowsRole reports whether the envelope permits the given actor role.
func (e *Envelope) AllowsRole(role string) bool {
	for _, r := range e.AllowedRoles {
		if r == role {
			return true
		}
	}
	return false
}

// AllowsWorkType reports whether the work type is within the envelope's plan.
func (e *Envelope) AllowsWorkType(workType string) bool {
	for _, wt := range e.Planning.AllowedWorkTypes {
		if wt == workType {
			return true
		}
	}
	return false
}

// SidecarPath returns the envelope sidecar path for a node under ticketsDir.
func SidecarPath(ticketsDir, nodeID string) string {
	return filepath.Join(ticketsDir, nodeID, "envelope.yaml")
}

// Write persists the envelope as its node's sidecar (the authoritative copy).
func Write(ticketsDir string, e *Envelope) error {
	dir := filepath.Join(ticketsDir, e.NodeID)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	data, err := yaml.Marshal(e)
	if err != nil {
		return err
	}
	return os.WriteFile(SidecarPath(ticketsDir, e.NodeID), data, 0o644)
}

// Read loads a node's envelope sidecar. The bool is false when no sidecar exists.
func Read(ticketsDir, nodeID string) (*Envelope, bool, error) {
	data, err := os.ReadFile(SidecarPath(ticketsDir, nodeID))
	if os.IsNotExist(err) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	var e Envelope
	if err := yaml.Unmarshal(data, &e); err != nil {
		return nil, false, err
	}
	return &e, true, nil
}
