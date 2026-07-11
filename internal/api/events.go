package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

func (s *Server) events(w http.ResponseWriter, r *http.Request, id string) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		writeError(w, http.StatusInternalServerError, "sse_unsupported", "streaming unsupported")
		return
	}
	updates, unsubscribe, ok := s.store.Subscribe(id)
	if !ok {
		writeError(w, http.StatusNotFound, "not_found", "job not found")
		return
	}
	defer unsubscribe()
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")
	heartbeat := time.NewTicker(15 * time.Second)
	defer heartbeat.Stop()
	for {
		select {
		case <-r.Context().Done():
			return
		case <-heartbeat.C:
			fmt.Fprint(w, ": keep-alive\n\n")
			flusher.Flush()
		case update := <-updates:
			payload, err := json.Marshal(update)
			if err != nil {
				return
			}
			fmt.Fprintf(w, "event: progress\ndata: %s\n\n", payload)
			flusher.Flush()
		}
	}
}
