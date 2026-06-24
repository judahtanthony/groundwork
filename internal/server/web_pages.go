package server

// Server-rendered operator pages (ADR 0042 progressive complexity: server HTML +
// SSE, no SPA) share one chrome — sidebar nav, topbar, and the SSE live-refresh
// script — defined once in web/layout.html.tmpl. Adding an operator page is three
// steps:
//
//  1. Add a content template web/<name>.content.tmpl that defines {{define
//     "content"}} and reads the page model under .Data.
//  2. Parse it together with the layout via newPage("web/<name>.content.tmpl")
//     into a package-level *template.Template.
//  3. Add a GET handler that builds a *pageView (chrome state + .Data) and calls
//     s.renderPage, then register its route in routes().
//
// The chrome reads pageView.Shell, pageView.Nav (active sidebar entry), and the
// breadcrumb fields; the page's own view model lives in pageView.Data. This keeps
// every operator screen live and visually consistent without duplicating layout.

import (
	"html/template"
	"net/http"
	"path/filepath"
)

// Active sidebar entries, matched against pageView.Nav for highlighting.
const (
	navDashboard = "dashboard"
	navTickets   = "tickets"
	navApprovals = "approvals"
)

// shell holds the chrome fields every page needs for the sidebar and topbar.
type shell struct {
	Repo, Branch, ServerAddr, DBSize, Version string
	PendingApprovals                          int
}

// pageView wraps a page's own model (Data) with the shared chrome state the
// layout template renders around it.
type pageView struct {
	Shell shell
	Nav   string // active sidebar entry (navDashboard/navTickets/navApprovals)
	Crumb string // breadcrumb section, e.g. "Operate"
	Title string // breadcrumb leaf and <title>
	Data  any    // page-specific view model, read as .Data in the content template
}

// newPage parses the shared layout together with one content template; the
// returned set renders a full page via its "layout" entry. It panics on parse
// errors (template bugs are programmer errors caught at startup), matching the
// existing dashboard template handling.
func newPage(contentFile string) *template.Template {
	return template.Must(template.ParseFS(webFS, "web/layout.html.tmpl", contentFile))
}

// renderPage executes tmpl's "layout" entry with pv. A mid-render error cannot be
// reported to the client (the header is already written), so it is dropped here
// as in handleDashboard.
func (s *Server) renderPage(w http.ResponseWriter, tmpl *template.Template, pv *pageView) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = tmpl.ExecuteTemplate(w, "layout", pv)
}

// shellState builds the chrome state shared by all pages. Callers that have
// already counted pending approvals pass the count to avoid a second query.
func (s *Server) shellState(pendingApprovals int) shell {
	return shell{
		Repo:             filepath.Base(s.proj.Root),
		Branch:           s.branch(),
		ServerAddr:       s.proj.Config.Server.Addr,
		DBSize:           dbSize(s.proj.DBPath()),
		Version:          s.version,
		PendingApprovals: pendingApprovals,
	}
}

// pendingCount returns the number of pending approvals for the chrome badge,
// tolerating store errors (the badge is non-critical).
func (s *Server) pendingCount() int {
	pending, err := s.db.ListApprovals("pending")
	if err != nil {
		return 0
	}
	return len(pending)
}

// placeholderTmpl renders a titled empty-state panel. Operator pages ship as live
// placeholders so the activated nav never 404s; each is replaced by its own ticket
// (Tickets by T-1063, Approvals by T-1062).
var placeholderTmpl = newPage("web/placeholder.content.tmpl")

type placeholder struct{ Heading, Message string }

// handleApprovalsPage serves the Approvals screen. Placeholder until T-1062.
func (s *Server) handleApprovalsPage(w http.ResponseWriter, r *http.Request) {
	s.renderPage(w, placeholderTmpl, &pageView{
		Shell: s.shellState(s.pendingCount()),
		Nav:   navApprovals,
		Crumb: "Operate",
		Title: "Approvals",
		Data:  placeholder{Heading: "Approvals", Message: "The approvals inbox lands with T-1062."},
	})
}
