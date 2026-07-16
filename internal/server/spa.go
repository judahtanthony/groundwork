package server

import (
	"net/http"

	webui "groundwork/web"
)

// spaHandler serves the Vite production build from the gw binary. Client-side
// routes use URL fragments, so the standard file server is sufficient and API
// and server-rendered operator routes remain independent.
var spaHandler = http.StripPrefix("/app/", http.FileServer(http.FS(webui.Dist)))

func (s *Server) handleSPAEntry(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/app/", http.StatusPermanentRedirect)
}
