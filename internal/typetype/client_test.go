package typetype

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNormalizeWatchURLUnwrapsTypeTypeWatchURL(t *testing.T) {
	input := "https://watch.eltux.fr/watch?v=https%3A%2F%2Fwww.youtube.com%2Fwatch%3Fv%3DdQw4w9WgXcQ"
	got, err := NormalizeWatchURL(input)
	if err != nil {
		t.Fatal(err)
	}
	want := "https://www.youtube.com/watch?v=dQw4w9WgXcQ"
	if got != want {
		t.Fatalf("NormalizeWatchURL() = %q, want %q", got, want)
	}
}

func TestNormalizeWatchURLKeepsDirectURL(t *testing.T) {
	input := "https://www.youtube.com/watch?v=dQw4w9WgXcQ"
	got, err := NormalizeWatchURL(input)
	if err != nil {
		t.Fatal(err)
	}
	if got != input {
		t.Fatalf("NormalizeWatchURL() = %q, want %q", got, input)
	}
}
func TestFetchStreamForwardsAuthorization(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got, want := r.URL.Path, "/streams/youtube/sabr"; got != want {
			t.Fatalf("path = %q, want %q", got, want)
		}
		if got, want := r.Header.Get("Authorization"), "Bearer user-token"; got != want {
			t.Fatalf("Authorization = %q, want %q", got, want)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"id","title":"Title","duration":1,"videoStreams":[],"videoOnlyStreams":[],"audioStreams":[],"hlsUrl":"","dashMpdUrl":""}`))
	}))
	defer server.Close()
	client := NewClient(server.URL)

	stream, err := client.FetchStream(t.Context(), "https://www.youtube.com/watch?v=abc123", "Bearer user-token")
	if err != nil {
		t.Fatal(err)
	}
	if stream.Title != "Title" {
		t.Fatalf("title = %q, want Title", stream.Title)
	}
}

func TestFetchStreamKeepsGenericEndpointForOtherProviders(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got, want := r.URL.Path, "/streams"; got != want {
			t.Fatalf("path = %q, want %q", got, want)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"id","title":"Title","duration":1,"videoStreams":[],"videoOnlyStreams":[],"audioStreams":[],"hlsUrl":"","dashMpdUrl":""}`))
	}))
	defer server.Close()
	client := NewClient(server.URL)

	if _, err := client.FetchStream(t.Context(), "https://www.nicovideo.jp/watch/sm9", ""); err != nil {
		t.Fatal(err)
	}
}

func TestClientMediaURLPreservesAPIBase(t *testing.T) {
	client := NewClient("http://server:8080/api")
	if got := client.MediaURL("/sabr/manifest/video"); got != "http://server:8080/api/sabr/manifest/video" {
		t.Fatalf("MediaURL() = %q", got)
	}
}

func TestProxyMediaURLUsesNicoManifestProxy(t *testing.T) {
	client := NewClient("http://server:8080")
	got := client.ProxyMediaURL("https://delivery.domand.nicovideo.jp/video.m3u8#cookie=value")
	if !strings.HasPrefix(got, "http://server:8080/proxy/nicovideo?url=") {
		t.Fatalf("url = %s", got)
	}
}

func TestProxyMediaURLUsesGeneralProxy(t *testing.T) {
	client := NewClient("http://server:8080")
	got := client.ProxyMediaURL("https://upos.example/video.m4s")
	if !strings.HasPrefix(got, "http://server:8080/proxy?url=") {
		t.Fatalf("url = %s", got)
	}
}
