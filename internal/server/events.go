package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// sseHeartbeat is how often a comment frame is sent to keep the connection (and
// any intermediary) alive when no events are flowing.
const sseHeartbeat = 15 * time.Second

// handleEvents streams coordinator events as Server-Sent Events (ADR 0025). It
// subscribes to the in-process hub and writes id/event/data frames with periodic
// heartbeats until the client disconnects. The hub keeps no replay buffer, so
// reconnection resumes from the live stream (Last-Event-ID is accepted but not
// replayed in v1).
func (s *Server) handleEvents(w http.ResponseWriter, r *http.Request) {
	if s.bus == nil {
		writeError(w, http.StatusServiceUnavailable, "events_unavailable", "coordinator event bus is not running")
		return
	}
	flusher, ok := w.(http.Flusher)
	if !ok {
		writeError(w, http.StatusInternalServerError, "streaming_unsupported", "response writer does not support streaming")
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.WriteHeader(http.StatusOK)

	ch, cancel := s.bus.Subscribe()
	defer cancel()

	fmt.Fprint(w, ": connected\n\n")
	flusher.Flush()

	ticker := time.NewTicker(sseHeartbeat)
	defer ticker.Stop()

	for {
		select {
		case <-r.Context().Done():
			return
		case ev, ok := <-ch:
			if !ok {
				return // hub closed
			}
			data, err := json.Marshal(ev)
			if err != nil {
				continue
			}
			// ev.Seq is the hub's monotonic id, so Last-Event-ID is a stable
			// offset a future replay buffer can resume from.
			fmt.Fprintf(w, "id: %d\nevent: %s\ndata: %s\n\n", ev.Seq, ev.Type, data)
			flusher.Flush()
		case <-ticker.C:
			fmt.Fprint(w, ": heartbeat\n\n")
			flusher.Flush()
		}
	}
}
