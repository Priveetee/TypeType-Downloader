package ffmpeg

import (
	"net/url"
	"slices"
	"strings"
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
	if query.Get("workload") != "download" {
		t.Fatalf("unexpected workload: %s", parsed.RawQuery)
	}
}

func TestSABRInputRetriesTransientHTTPFailures(t *testing.T) {
	args := sabrInputArgs()
	for _, value := range []string{"-reconnect", "-reconnect_streamed", "-reconnect_on_network_error", "-reconnect_on_http_error", "4xx,5xx"} {
		if !slices.Contains(args, value) {
			t.Fatalf("missing ffmpeg retry argument %q in %v", value, args)
		}
	}
	if slices.Contains(args, "-reconnect_at_eof") {
		t.Fatalf("static SABR manifests must not reconnect at EOF: %v", args)
	}
}

func TestReadProgressPublishesChangedTotalSize(t *testing.T) {
	values := make([]int64, 0, 2)
	err := readProgress(strings.NewReader("frame=1\ntotal_size=1024\nprogress=continue\ntotal_size=1024\ntotal_size=4096\n"), func(downloadedBytes int64) {
		values = append(values, downloadedBytes)
	})
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(values, []int64{1024, 4096}) {
		t.Fatalf("progress = %v, want [1024 4096]", values)
	}
}
