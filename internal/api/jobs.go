package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"typetype-downloader-go/internal/artifact"
	"typetype-downloader-go/internal/job"
	"typetype-downloader-go/internal/typetype"
)

func (s *Server) createJob(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var request job.CreateRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "invalid JSON body")
		return
	}
	request.URL = strings.TrimSpace(request.URL)
	if request.URL == "" {
		writeError(w, http.StatusBadRequest, "invalid_url", "url is required")
		return
	}
	normalized, err := typetype.NormalizeWatchURL(request.URL)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_url", err.Error())
		return
	}
	record, cached, created, err := s.store.Create(normalized, request.Options, r.Header.Get("Authorization"))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "create_failed", err.Error())
		return
	}
	if created {
		if err := s.runner.Enqueue(record.ID); err != nil {
			s.store.Fail(record.ID, "queue_full", err)
			writeError(w, http.StatusTooManyRequests, "queue_full", err.Error())
			return
		}
	}
	writeJSON(w, http.StatusCreated, job.CreateResponse{ID: record.ID, Cached: cached})
}

func (s *Server) getJob(w http.ResponseWriter, id string) {
	response, ok := s.store.Response(id)
	if !ok {
		writeError(w, http.StatusNotFound, "not_found", "job not found")
		return
	}
	writeJSON(w, http.StatusOK, response)
}

func (s *Server) cancel(w http.ResponseWriter, id string) {
	if !s.store.Cancel(id) {
		if _, ok := s.store.Get(id); ok {
			writeError(w, http.StatusConflict, "not_cancellable", "job is not cancellable")
			return
		}
		writeError(w, http.StatusNotFound, "not_found", "job not found")
		return
	}
	writeJSON(w, http.StatusAccepted, map[string]string{"status": "cancelled"})
}

func (s *Server) deleteJob(w http.ResponseWriter, r *http.Request, id string) {
	record, ok := s.store.Delete(id)
	if !ok {
		if record, exists := s.store.Get(id); exists && record.Status == job.StatusRunning {
			writeError(w, http.StatusConflict, "running", "job is running")
			return
		}
		writeError(w, http.StatusNotFound, "not_found", "job not found")
		return
	}
	if record.Artifact != "" {
		_ = s.files.Delete(r.Context(), artifact.Saved{Backend: record.Storage, Location: record.Artifact})
	}
	w.WriteHeader(http.StatusNoContent)
}
