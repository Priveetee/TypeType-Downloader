package sabr

import (
	"net/url"
	"strings"
	"testing"
)

func TestBuildManifestURLIncludesSelectedTracks(t *testing.T) {
	result, err := buildManifestURL(Options{
		ManifestURL: "http://server/api/sabr/manifest/video",
		VideoItag:   137, AudioItag: 140, AudioTrackID: "fr-FR.4",
	})
	if err != nil {
		t.Fatal(err)
	}
	parsed, err := url.Parse(result)
	if err != nil {
		t.Fatal(err)
	}
	query := parsed.Query()
	if query.Get("videoItag") != "137" || query.Get("audioItag") != "140" || query.Get("audioTrackId") != "fr-FR.4" {
		t.Fatalf("unexpected query: %s", parsed.RawQuery)
	}
	if query.Get("workload") != "download" {
		t.Fatalf("unexpected workload: %s", parsed.RawQuery)
	}
}

func TestParseManifestResolvesRelativeURLs(t *testing.T) {
	base, err := url.Parse("https://server/api/sabr/manifest/video")
	if err != nil {
		t.Fatal(err)
	}
	tracks, err := parseManifest(strings.NewReader(testManifest(false)), base, false)
	if err != nil {
		t.Fatal(err)
	}
	if tracks[0].URLs[1] != "https://server/video/1" || tracks[1].URLs[2] != "https://server/audio/2" {
		t.Fatalf("unexpected tracks: %+v", tracks)
	}
}
