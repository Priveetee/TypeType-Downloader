package api

import (
	"context"
	"net/http"
	"strings"
	"time"

	"typetype-downloader-go/internal/artifact"
	"typetype-downloader-go/internal/job"
	"typetype-downloader-go/internal/pipeline"
	"typetype-downloader-go/internal/storage"
)

type Server struct {
	store  *job.Store
	runner *pipeline.Runner
	files  artifact.Store
	disk   *storage.Monitor
	checks []HealthCheck
}

type HealthCheck interface {
	Name() string
	Health(context.Context) error
}

func NewServer(store *job.Store, runner *pipeline.Runner, files artifact.Store, disk *storage.Monitor, health ...HealthCheck) *Server {
	return &Server{store: store, runner: runner, files: files, disk: disk, checks: health}
}

func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", s.health)
	mux.HandleFunc("/health/deep", s.healthDeep)
	mux.HandleFunc("/jobs", s.jobs)
	mux.HandleFunc("/jobs/", s.jobByID)
	return mux
}

func (s *Server) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "service": "typetype-downloader"})
}

func (s *Server) healthDeep(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()
	checks := map[string]string{"service": "ok"}
	status := http.StatusOK
	for _, check := range s.checks {
		if err := check.Health(ctx); err != nil {
			checks[check.Name()] = err.Error()
			status = http.StatusServiceUnavailable
		} else {
			checks[check.Name()] = "ok"
		}
	}
	if err := s.disk.Health(); err != nil {
		checks[s.disk.Name()] = err.Error()
		status = http.StatusServiceUnavailable
	} else {
		checks[s.disk.Name()] = "ok"
	}
	writeJSON(w, status, map[string]any{"status": statusText(status), "checks": checks})
}

func statusText(status int) string {
	if status == http.StatusOK {
		return "ok"
	}
	return "degraded"
}

func (s *Server) jobs(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/jobs" {
		writeError(w, http.StatusNotFound, "not_found", "not found")
		return
	}
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed")
		return
	}
	s.createJob(w, r)
}

func (s *Server) jobByID(w http.ResponseWriter, r *http.Request) {
	id, action, ok := parseJobPath(r.URL.Path)
	if !ok {
		writeError(w, http.StatusNotFound, "not_found", "not found")
		return
	}
	if action == "" && r.Method == http.MethodGet {
		s.getJob(w, id)
		return
	}
	if action == "" && r.Method == http.MethodDelete {
		s.deleteJob(w, r, id)
		return
	}
	if action == "events" && r.Method == http.MethodGet {
		s.events(w, r, id)
		return
	}
	if action == "artifact" && r.Method == http.MethodGet {
		s.artifact(w, r, id)
		return
	}
	if action == "cancel" && r.Method == http.MethodPost {
		s.cancel(w, id)
		return
	}
	writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed")
}

func parseJobPath(path string) (string, string, bool) {
	path = strings.TrimPrefix(path, "/jobs/")
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		return "", "", false
	}
	if len(parts) == 1 {
		return parts[0], "", true
	}
	if len(parts) == 2 {
		return parts[0], parts[1], true
	}
	return "", "", false
}
