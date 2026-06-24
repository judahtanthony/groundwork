package server

// Approvals inbox (T-1062): the operator's read of pending decisions, grouped by
// risk class and annotated with the requesting actor, the required actor/role
// constraints, the ticket context, and a derived gate reason explaining why the
// action is held for a human (ADR 0028). This screen is read-only; the
// approve/reject/clarify actions are wired by T-1064.

import (
	"net/http"
	"strconv"
	"strings"

	"groundwork/internal/approval"
	"groundwork/internal/store/sqlite"
)

var approvalsTmpl = newPage("web/approvals.content.tmpl")

// approvalItem is one pending approval rendered with its full decision context.
type approvalItem struct {
	ID, Type, Summary                                      string
	RiskClass, RiskScore, Reversible                       string
	RequestedBy, RequiredActors, RequiredRoles, GateReason string
	TicketID, TicketTitle, TicketStatus, TicketWorkType    string
	Created                                                string
}

// approvalGroup buckets pending approvals under one risk class.
type approvalGroup struct {
	Risk  string
	Items []approvalItem
}

type approvalsData struct {
	Groups []approvalGroup
	Total  int
}

// riskOrder ranks risk classes most-severe first for the grouped inbox.
var riskOrder = []string{"critical", "high", "medium", "low"}

func (s *Server) handleApprovalsPage(w http.ResponseWriter, r *http.Request) {
	data, err := s.buildApprovals()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "approvals_failed", err.Error())
		return
	}
	s.renderPage(w, approvalsTmpl, &pageView{
		Shell: s.shellState(data.Total),
		Nav:   navApprovals,
		Crumb: "Operate",
		Title: "Approvals",
		Data:  data,
	})
}

// handleApprovalDecideForm handles the inbox's approve/reject/clarify form posts
// (T-1064). It routes through the same ApprovalService path as the JSON API and
// the CLI — the UI never self-approves or bypasses the gate (ADR 0028) — then
// redirects back to the inbox (POST-redirect-GET) so the SSE-refreshed page shows
// the result.
func (s *Server) handleApprovalDecideForm(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		writeError(w, http.StatusBadRequest, "bad_form", err.Error())
		return
	}
	to, ok := decisionStatus(r.PostFormValue("decision"))
	if !ok {
		writeError(w, http.StatusBadRequest, "bad_decision", "decision must be approve, reject, or clarify")
		return
	}
	if _, ok := s.applyDecision(w, r.PathValue("id"), to, r.PostFormValue("reason")); !ok {
		return
	}
	http.Redirect(w, r, "/approvals", http.StatusSeeOther)
}

// decisionStatus maps an inbox button value to an approval status.
func decisionStatus(v string) (approval.Status, bool) {
	switch v {
	case "approve":
		return approval.StatusApproved, true
	case "reject":
		return approval.StatusRejected, true
	case "clarify":
		return approval.StatusClarifying, true
	}
	return "", false
}

func (s *Server) buildApprovals() (*approvalsData, error) {
	pending, err := s.db.ListApprovals(string(approval.StatusPending))
	if err != nil {
		return nil, err
	}
	all, err := s.db.ListTickets()
	if err != nil {
		return nil, err
	}
	titles := make(map[string]string, len(all))
	statuses := make(map[string]string, len(all))
	workTypes := make(map[string]string, len(all))
	for _, t := range all {
		titles[t.ID] = t.Title
		statuses[t.ID] = string(t.Status)
		workTypes[t.ID] = t.WorkType
	}

	grouped := map[string][]approvalItem{}
	for _, a := range pending {
		risk := a.RiskClass
		if risk == "" {
			risk = "unclassified"
		}
		grouped[risk] = append(grouped[risk], approvalItem{
			ID: a.ID, Type: a.Type, Summary: orDash(a.Summary),
			RiskClass: risk, RiskScore: intPtrStr(a.RiskScore), Reversible: boolPtrStr(a.Reversible),
			RequestedBy:    orDash(a.RequestedByActor),
			RequiredActors: orList(a.RequiredActors),
			RequiredRoles:  orList(a.RequiredRoles),
			GateReason:     gateReason(a),
			TicketID:       a.TicketID,
			TicketTitle:    orDash(titles[a.TicketID]),
			TicketStatus:   orDash(statuses[a.TicketID]),
			TicketWorkType: orDash(workTypes[a.TicketID]),
			Created:        relTime(a.CreatedAt),
		})
	}

	d := &approvalsData{Total: len(pending)}
	// Known risk classes in severity order, then any others (e.g. unclassified).
	emitted := map[string]bool{}
	for _, risk := range riskOrder {
		if items := grouped[risk]; len(items) > 0 {
			d.Groups = append(d.Groups, approvalGroup{Risk: risk, Items: items})
			emitted[risk] = true
		}
	}
	for risk, items := range grouped {
		if !emitted[risk] {
			d.Groups = append(d.Groups, approvalGroup{Risk: risk, Items: items})
		}
	}
	return d, nil
}

// gateReason explains, for a pending approval, why it is held for a decision: an
// explicit actor/role constraint, the v1 human gate for the type, or a plain
// policy review (ADR 0028).
func gateReason(a *sqlite.Approval) string {
	switch {
	case len(a.RequiredActors) > 0:
		return "requires actor: " + strings.Join(a.RequiredActors, ", ")
	case len(a.RequiredRoles) > 0:
		return "requires role: " + strings.Join(a.RequiredRoles, ", ")
	case humanGated(approval.Type(a.Type)):
		return "human-gated (" + a.Type + ")"
	default:
		return "policy review"
	}
}

func intPtrStr(p *int) string {
	if p == nil {
		return "—"
	}
	return strconv.Itoa(*p)
}

func boolPtrStr(p *bool) string {
	if p == nil {
		return "—"
	}
	if *p {
		return "reversible"
	}
	return "irreversible"
}

func orList(xs []string) string {
	if len(xs) == 0 {
		return "—"
	}
	return strings.Join(xs, ", ")
}
