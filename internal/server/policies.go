package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"sort"

	"gopkg.in/yaml.v3"
	"groundwork/internal/approval"
	"groundwork/internal/config"
	"groundwork/internal/policy"
	"groundwork/internal/risk"
	"groundwork/internal/store/sqlite"
)

type policyRuleView struct {
	Group string      `json:"group"`
	Order int         `json:"order"`
	Rule  policy.Rule `json:"rule"`
}

type validationTemplateView struct {
	Name     string                    `json:"name"`
	Template policy.ValidationTemplate `json:"template"`
}

type policiesResponse struct {
	Trust               *policy.TrustPolicy      `json:"trust,omitempty"`
	Rules               []policyRuleView         `json:"rules"`
	Validation          *policy.ValidationPolicy `json:"validation,omitempty"`
	ValidationTemplates []validationTemplateView `json:"validation_templates"`
	Warnings            []string                 `json:"warnings"`
}

type policyAmendment struct {
	Trust policy.TrustPolicy `json:"trust"`
}

type policyUpdateRequest struct {
	TicketID string             `json:"ticket_id"`
	Trust    policy.TrustPolicy `json:"trust"`
}

func (s *Server) handlePolicies(w http.ResponseWriter, _ *http.Request) {
	set, warnings, err := policy.Load(s.proj.PoliciesDir())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "policy_load_failed", err.Error())
		return
	}
	view := policiesResponse{Trust: set.Trust, Validation: set.Validation, Warnings: warnings}
	if set.Trust != nil {
		appendRules := func(group string, rules []policy.Rule) {
			for i, rule := range rules {
				view.Rules = append(view.Rules, policyRuleView{Group: group, Order: i + 1, Rule: rule})
			}
		}
		appendRules("require_human", set.Trust.RequireHuman)
		appendRules("auto_approve", set.Trust.AutoApprove)
		appendRules("allow_claim", set.Trust.AllowClaim)
	}
	if set.Validation != nil {
		names := make([]string, 0, len(set.Validation.Templates))
		for name := range set.Validation.Templates {
			names = append(names, name)
		}
		sort.Strings(names)
		for _, name := range names {
			view.ValidationTemplates = append(view.ValidationTemplates, validationTemplateView{Name: name, Template: set.Validation.Templates[name]})
		}
	}
	writeJSON(w, http.StatusOK, view)
}

func (s *Server) handlePolicyUpdate(w http.ResponseWriter, r *http.Request) {
	if s.approvals == nil {
		writeError(w, http.StatusServiceUnavailable, "approvals_unavailable", errApprovalsUnavailable.Error())
		return
	}
	var body policyUpdateRequest
	if !decodeJSON(w, r, &body) {
		return
	}
	if body.TicketID == "" {
		writeError(w, http.StatusBadRequest, "invalid_request", "ticket_id is required for an auditable policy amendment")
		return
	}
	if _, err := s.db.GetTicket(body.TicketID); err != nil {
		s.writeStoreError(w, err)
		return
	}
	if err := s.validateTrustReplacement(&body.Trust); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_policy", err.Error())
		return
	}
	actionJSON, err := json.Marshal(policyAmendment{Trust: body.Trust})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "encode_failed", err.Error())
		return
	}
	owner, _ := s.approvals.registry.Resolve(ownerActor)
	appr, err := s.approvals.Request(RequestParams{
		TicketID: body.TicketID,
		Type:     approval.TypeAmendPolicy,
		Summary:  "Amend ordered trust policy rules",
		Action: policy.Action{
			Type:  string(approval.TypeAmendPolicy),
			Actor: owner,
			Scope: risk.Scope{Files: []string{filepath.ToSlash(filepath.Join(config.GroundworkDir, "policies", "trust.yaml"))}},
		},
		ActionJSON: string(actionJSON),
	})
	if err != nil {
		s.writeMutationError(w, err)
		return
	}
	if appr.Status == string(approval.StatusApproved) {
		if err := s.applyPolicyAmendment(appr.ActionJSON); err != nil {
			writeError(w, http.StatusInternalServerError, "policy_write_failed", err.Error())
			return
		}
	}
	writeJSON(w, http.StatusAccepted, map[string]any{"approval": appr})
}

func (s *Server) validateTrustReplacement(next *policy.TrustPolicy) error {
	data, err := yaml.Marshal(next)
	if err != nil {
		return err
	}
	parsed, _, err := policy.ParseTrust(data)
	if err != nil {
		return err
	}
	current, _, err := policy.Load(s.proj.PoliciesDir())
	if err != nil {
		return err
	}
	if current.Trust == nil {
		return errors.New("no trust policy is configured")
	}
	ids := func(rules []policy.Rule) []string {
		out := make([]string, len(rules))
		for i, rule := range rules {
			out[i] = rule.ID
		}
		return out
	}
	groups := []struct {
		name          string
		before, after []policy.Rule
	}{
		{"require_human", current.Trust.RequireHuman, parsed.RequireHuman},
		{"auto_approve", current.Trust.AutoApprove, parsed.AutoApprove},
		{"allow_claim", current.Trust.AllowClaim, parsed.AllowClaim},
	}
	for _, group := range groups {
		before, after := ids(group.before), ids(group.after)
		if !slices.Equal(before, after) {
			return fmt.Errorf("%s trust rule stable ids and ordering must be preserved (want %v, got %v)", group.name, before, after)
		}
	}
	return nil
}

func (s *Server) applyPolicyAmendment(raw string) error {
	var amendment policyAmendment
	if err := json.Unmarshal([]byte(raw), &amendment); err != nil {
		return fmt.Errorf("decode policy amendment: %w", err)
	}
	if err := s.validateTrustReplacement(&amendment.Trust); err != nil {
		return err
	}
	data, err := yaml.Marshal(&amendment.Trust)
	if err != nil {
		return err
	}
	path := filepath.Join(s.proj.PoliciesDir(), "trust.yaml")
	tmp, err := os.CreateTemp(s.proj.PoliciesDir(), ".trust-*.yaml")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	defer os.Remove(tmpPath)
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Chmod(0o644); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Rename(tmpPath, path); err != nil {
		return err
	}
	s.approvals.policies.ReplaceTrust(&amendment.Trust)
	return nil
}

func (s *Server) handlePolicySuggestions(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")
	if status == "" {
		status = "pending"
	}
	if status != "pending" && status != "all" {
		writeError(w, http.StatusBadRequest, "invalid_status", "status must be pending or all")
		return
	}
	if status == "all" {
		status = ""
	}
	items, err := s.db.ListSuggestions(status)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "list_failed", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, items)
}

func (s *Server) handlePolicySuggestionPromote(w http.ResponseWriter, r *http.Request) {
	s.handlePolicySuggestionDecision(w, r, "promoted")
}

func (s *Server) handlePolicySuggestionDismiss(w http.ResponseWriter, r *http.Request) {
	s.handlePolicySuggestionDecision(w, r, "dismissed")
}

func (s *Server) handlePolicySuggestionDecision(w http.ResponseWriter, r *http.Request, status string) {
	item, err := s.db.SetSuggestionStatus(r.PathValue("id"), status)
	if err != nil {
		if errors.Is(err, sqlite.ErrNotFound) {
			writeError(w, http.StatusNotFound, "not_found", "policy suggestion not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "store_error", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, item)
}
