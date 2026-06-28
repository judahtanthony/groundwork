package server

// Parent/root review screen (T-1086, ADR 0057): renders the feature-level review
// bundle on the Phase 4 server-rendered operator UI so a human reviews a subtree
// once — per-leaf summary, validation, and exceptions, with a recommendation —
// to inform the root land_to_main decision.

import "net/http"

var reviewTmpl = newPage("web/review.content.tmpl")

func (s *Server) handleReviewPage(w http.ResponseWriter, r *http.Request) {
	b, err := s.db.ReviewBundle(r.PathValue("id"))
	if err != nil {
		s.writeStoreError(w, err)
		return
	}
	s.renderPage(w, reviewTmpl, &pageView{
		Shell: s.shellState(s.pendingCount()),
		Nav:   navTickets,
		Crumb: "Operate / Review",
		Title: b.NodeID,
		Data:  b,
	})
}
