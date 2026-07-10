package ffmpeg

import (
	"net/url"
	"testing"
)

func TestSABRURLIncludesSelectedTracks(t *testing.T) {
	result, err := sabrURL(SABROptions{
		ManifestURL:  "http://server/api/sabr/manifest/video",
		VideoItag:    137,
		AudioItag:    140,
		AudioTrackID: "fr-FR.4",
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
}
