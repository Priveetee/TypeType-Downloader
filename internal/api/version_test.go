package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"typetype-downloader-go/internal/buildinfo"
)

func TestVersionReturnsBuildMetadata(t *testing.T) {
	server := &Server{}
	request := httptest.NewRequest(http.MethodGet, "/version", nil)
	response := httptest.NewRecorder()
	server.Routes().ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d", response.Code)
	}
	var body map[string]string
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body["revision"] != buildinfo.Revision || body["service"] != "downloader" {
		t.Fatalf("body = %#v", body)
	}
}
