package api

import (
	"encoding/json"
	"math"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"typetype-downloader-go/internal/storage"
)

func unavailableDisk(t *testing.T) *storage.Monitor {
	t.Helper()
	monitor, err := storage.NewMonitor(t.TempDir(), math.MaxInt64, 20)
	if err != nil {
		t.Fatal(err)
	}
	return monitor
}

func TestCreateJobRejectsInsufficientStorage(t *testing.T) {
	server := &Server{disk: unavailableDisk(t)}
	request := httptest.NewRequest(http.MethodPost, "/jobs", strings.NewReader(`{"url":"https://youtu.be/test"}`))
	response := httptest.NewRecorder()

	server.Routes().ServeHTTP(response, request)

	if response.Code != http.StatusInsufficientStorage {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
	}
	var body map[string]any
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body["code"] != "insufficient_storage" || body["errorCode"] != "insufficient_storage" {
		t.Fatalf("body = %#v", body)
	}
}

func TestDeepHealthReportsDegradedDisk(t *testing.T) {
	server := &Server{disk: unavailableDisk(t)}
	request := httptest.NewRequest(http.MethodGet, "/health/deep", nil)
	response := httptest.NewRecorder()

	server.Routes().ServeHTTP(response, request)

	if response.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
	}
	if !strings.Contains(response.Body.String(), `"disk":"free bytes`) {
		t.Fatalf("body = %s", response.Body.String())
	}
}
