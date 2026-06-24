package server

import (
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"groundwork/internal/approval"
	"groundwork/internal/run"
	"groundwork/internal/store/sqlite"
	"groundwork/internal/ticket"
)

//go:embed web/dashboard.html.tmpl web/groundwork.css
var webFS embed.FS

var dashboardTmpl = template.Must(template.ParseFS(webFS, "web/dashboard.html.tmpl"))

// --- view models ---

type kpi struct{ Label, Value, Sub, Tone string }
type runRow struct{ Title, Ticket, RunID, Actor, Mode, Status, Elapsed, Last, Tone string }
type attnItem struct{ Tone, Tag, Title, Meta string }
type eventRow struct{ Time, Event, Object, Who string }
type statusCount struct {
	Status string
	Count  int
}

type dashboardData struct {
	Repo, Branch, ServerAddr, DBSize, Version, Now string
	Runtime, Sandbox, Uptime                       string
	Concurrency                                    int
	PendingApprovals                               int
	KPIs                                           []kpi
	Runs                                           []runRow
	Attention                                      []attnItem
	Events                                         []eventRow
	StatusBreakdown                                []statusCount
}

// handleDashboard renders the server-rendered operations dashboard (T-0801): KPI
// counts, active runs, the attention queue, and a recent-event timeline over
// coordinator state.
func (s *Server) handleDashboard(w http.ResponseWriter, r *http.Request) {
	data, err := s.buildDashboard()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "dashboard_failed", err.Error())
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := dashboardTmpl.Execute(w, data); err != nil {
		// Header is already written; nothing actionable beyond logging upstream.
		return
	}
}

// handleDashboardCSS serves the embedded stylesheet.
func (s *Server) handleDashboardCSS(w http.ResponseWriter, r *http.Request) {
	b, err := webFS.ReadFile("web/groundwork.css")
	if err != nil {
		writeError(w, http.StatusInternalServerError, "asset_missing", err.Error())
		return
	}
	w.Header().Set("Content-Type", "text/css; charset=utf-8")
	w.Write(b)
}

func (s *Server) buildDashboard() (*dashboardData, error) {
	all, err := s.db.ListTickets()
	if err != nil {
		return nil, err
	}
	byID := make(map[string]*ticket.Ticket, len(all))
	counts := map[ticket.Status]int{}
	for _, t := range all {
		byID[t.ID] = t
		counts[t.Status]++
	}

	eligible, blocked := 0, 0
	var blockedNodes []*ticket.Ticket
	for _, t := range all {
		if t.Status != ticket.StatusTodo {
			continue
		}
		ok, err := s.db.DependenciesSatisfied(t.ID)
		if err != nil {
			return nil, err
		}
		if ok {
			eligible++
		} else {
			blocked++
			blockedNodes = append(blockedNodes, t)
		}
	}

	pending, err := s.db.ListApprovals(string(approval.StatusPending))
	if err != nil {
		return nil, err
	}
	allRuns, err := s.db.ListRuns()
	if err != nil {
		return nil, err
	}
	events, err := s.db.RecentAuditEvents(50)
	if err != nil {
		return nil, err
	}

	d := &dashboardData{
		Repo:             filepath.Base(s.proj.Root),
		Branch:           s.branch(),
		ServerAddr:       s.proj.Config.Server.Addr,
		Version:          s.version,
		Now:              time.Now().Format("15:04:05"),
		Runtime:          orDash(s.proj.Config.Runtime),
		Sandbox:          orDash(s.proj.Config.Sandbox),
		Concurrency:      s.proj.Config.MaxConcurrency,
		Uptime:           humanDuration(time.Since(s.started)),
		DBSize:           dbSize(s.proj.DBPath()),
		PendingApprovals: len(pending),
	}

	// Active runs.
	active := activeRuns(allRuns)
	for _, rn := range active {
		title := rn.TicketID
		if t := byID[rn.TicketID]; t != nil {
			title = t.Title
		}
		d.Runs = append(d.Runs, runRow{
			Title: title, Ticket: rn.TicketID, RunID: rn.ID,
			Actor: orDash(rn.ActorID), Mode: orDash(rn.Mode),
			Status: rn.Status, Tone: runTone(rn.Status),
			Elapsed: elapsedSince(rn.StartedAt), Last: lastEvent(rn),
		})
	}

	// KPIs.
	d.KPIs = []kpi{
		{Label: "Active runs", Value: itoa(len(active)), Tone: toneIf(len(active) > 0, "run")},
		{Label: "In review", Value: itoa(counts[ticket.StatusReview]), Tone: toneIf(counts[ticket.StatusReview] > 0, "run")},
		{Label: "Blocked", Value: itoa(blocked), Tone: toneIf(blocked > 0, "bad")},
		{Label: "Pending approvals", Value: itoa(len(pending)), Tone: toneIf(len(pending) > 0, "warn")},
		{Label: "Ready", Value: itoa(eligible), Sub: "eligible to start", Tone: toneIf(eligible > 0, "ok")},
		{Label: "Landed today", Value: itoa(landedToday(events)), Tone: "ok"},
	}

	// Attention queue: pending approvals first, then blocked work.
	for _, a := range pending {
		d.Attention = append(d.Attention, attnItem{
			Tone: "warn", Tag: "approval",
			Title: orDash(a.Summary),
			Meta:  fmt.Sprintf("%s · risk %s · %s", a.Type, orDash(a.RiskClass), a.TicketID),
		})
	}
	for _, t := range blockedNodes {
		d.Attention = append(d.Attention, attnItem{
			Tone: "bad", Tag: "blocked",
			Title: t.Title,
			Meta:  fmt.Sprintf("%s · blocked by %s", t.ID, strings.Join(s.unmetDepIDs(t.ID), ", ")),
		})
	}

	// Recent-event timeline (newest first, capped).
	for i, e := range events {
		if i >= 18 {
			break
		}
		d.Events = append(d.Events, eventRow{
			Time:   relTime(e.CreatedAt),
			Event:  eventLabel(e),
			Object: e.ObjectID,
			Who:    orDash(e.Actor),
		})
	}

	// Work-tree status breakdown, canonical order, non-zero only.
	for _, st := range ticket.AllStatuses {
		if n := counts[st]; n > 0 {
			d.StatusBreakdown = append(d.StatusBreakdown, statusCount{Status: string(st), Count: n})
		}
	}

	return d, nil
}

// --- helpers ---

func (s *Server) branch() string {
	if s.repo == nil {
		return "—"
	}
	b, err := s.repo.CurrentBranch()
	if err != nil || b == "" {
		return "—"
	}
	return b
}

// unmetDepIDs returns the ids of id's dependencies that are not yet done.
func (s *Server) unmetDepIDs(id string) []string {
	depIDs, err := s.db.DependencyIDs(id)
	if err != nil {
		return nil
	}
	var out []string
	for _, depID := range depIDs {
		dep, err := s.db.GetTicket(depID)
		if err != nil {
			continue
		}
		if !ticket.DependencyMet(dep.Status) {
			out = append(out, depID)
		}
	}
	return out
}

func activeRuns(runs []*sqlite.Run) []*sqlite.Run {
	var out []*sqlite.Run
	for _, rn := range runs {
		switch run.Status(rn.Status) {
		case run.StatusRunning, run.StatusPending, run.StatusPaused:
			out = append(out, rn)
		}
	}
	return out
}

func runTone(status string) string {
	switch run.Status(status) {
	case run.StatusRunning:
		return "run"
	case run.StatusPaused:
		return "warn"
	default:
		return "idle"
	}
}

func lastEvent(rn *sqlite.Run) string {
	if rn.LastMessage != "" {
		return rn.LastMessage
	}
	return orDash(rn.LastEvent)
}

// landedToday counts ticket.landed audit events dated today (UTC).
func landedToday(events []sqlite.AuditEvent) int {
	today := time.Now().UTC().Format("2006-01-02")
	n := 0
	for _, e := range events {
		if e.Type == "ticket.landed" && strings.HasPrefix(e.CreatedAt, today) {
			n++
		}
	}
	return n
}

// eventLabel formats an audit event into a short verb, appending the target
// status for transitions/landings when the payload carries it.
func eventLabel(e sqlite.AuditEvent) string {
	verb := e.Type
	if i := strings.LastIndex(verb, "."); i >= 0 {
		verb = verb[i+1:]
	}
	if e.Payload != "" {
		var p map[string]any
		if json.Unmarshal([]byte(e.Payload), &p) == nil {
			if to, ok := p["to"].(string); ok && to != "" {
				return verb + " → " + to
			}
		}
	}
	return verb
}

func relTime(ts string) string {
	t, err := time.Parse(time.RFC3339, ts)
	if err != nil {
		return "—"
	}
	d := time.Since(t)
	switch {
	case d < time.Minute:
		return "now"
	case d < time.Hour:
		return fmt.Sprintf("%dm", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh", int(d.Hours()))
	default:
		return fmt.Sprintf("%dd", int(d.Hours()/24))
	}
}

func elapsedSince(ts string) string {
	t, err := time.Parse(time.RFC3339, ts)
	if err != nil {
		return "—"
	}
	d := time.Since(t)
	if d < 0 {
		d = 0
	}
	return humanDuration(d)
}

func humanDuration(d time.Duration) string {
	d = d.Round(time.Second)
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	sec := int(d.Seconds()) % 60
	switch {
	case h > 0:
		return fmt.Sprintf("%dh %dm", h, m)
	case m > 0:
		return fmt.Sprintf("%dm %ds", m, sec)
	default:
		return fmt.Sprintf("%ds", sec)
	}
}

func dbSize(path string) string {
	fi, err := os.Stat(path)
	if err != nil {
		return "—"
	}
	return humanBytes(fi.Size())
}

func humanBytes(n int64) string {
	const unit = 1024
	if n < unit {
		return fmt.Sprintf("%d B", n)
	}
	div, exp := int64(unit), 0
	for x := n / unit; x >= unit; x /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(n)/float64(div), "KMGT"[exp])
}

func toneIf(cond bool, tone string) string {
	if cond {
		return tone
	}
	return "idle"
}

func orDash(s string) string {
	if s == "" {
		return "—"
	}
	return s
}

func itoa(n int) string { return fmt.Sprintf("%d", n) }
