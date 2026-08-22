package artifact

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestS3StoreMarksInternalRedirect(t *testing.T) {
	store := newTestS3Store(t, "typetype-garage:3900")
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/artifact", nil)

	if err := store.ServeHTTP(response, request, Saved{Location: "artifact.mp4"}, "video.mp4"); err != nil {
		t.Fatal(err)
	}

	if response.Code != http.StatusFound {
		t.Fatalf("status = %d", response.Code)
	}
	if response.Header().Get(internalRedirectHeader) != "1" {
		t.Fatalf("internal redirect header = %q", response.Header().Get(internalRedirectHeader))
	}
}

func TestS3StoreLeavesPublicRedirectUnmarked(t *testing.T) {
	store := newTestS3Store(t, "downloads.example.com")
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/artifact", nil)

	if err := store.ServeHTTP(response, request, Saved{Location: "artifact.mp4"}, "video.mp4"); err != nil {
		t.Fatal(err)
	}

	if response.Header().Get(internalRedirectHeader) != "" {
		t.Fatalf("public redirect was marked internal")
	}
}

func newTestS3Store(t *testing.T, publicEndpoint string) *S3Store {
	t.Helper()
	store, err := NewS3Store(S3Config{
		Endpoint:       "typetype-garage:3900",
		PublicEndpoint: publicEndpoint,
		Region:         "garage",
		Bucket:         "downloads",
		AccessKey:      "key",
		SecretKey:      "secret",
		PathStyle:      true,
	})
	if err != nil {
		t.Fatal(err)
	}
	return store
}
