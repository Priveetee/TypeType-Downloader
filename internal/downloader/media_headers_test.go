package downloader

import (
	"net/http"
	"testing"
)

func TestEffectiveRangeModeForcesHeaderForProxyAndProviderMedia(t *testing.T) {
	tests := []struct {
		name string
		url  string
	}{
		{"server proxy", "http://typetype-server:8080/proxy?url=https%3A%2F%2Fexample.com%2Fv.mp4"},
		{"nico proxy", "http://typetype-server:8080/proxy/nicovideo?url=https%3A%2F%2Fexample.com%2Fv.ts"},
		{"bilibili", "https://upos-hz-mirrorakam.akamaized.net/video.m4s?deadline=100"},
		{"nico", "https://delivery.domand.nicovideo.jp/hlsbid/abc/seg.ts"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := effectiveRangeMode(test.url, "query"); got != "header" {
				t.Fatalf("effectiveRangeMode() = %q, want header", got)
			}
		})
	}
}

func TestEffectiveRangeModeKeepsConfiguredModeForGoogleVideo(t *testing.T) {
	rawURL := "https://rr1---sn.googlevideo.com/videoplayback?id=x"
	if got := effectiveRangeMode(rawURL, "query"); got != "query" {
		t.Fatalf("effectiveRangeMode() = %q, want query", got)
	}
}

func TestApplyMediaHeadersForBilibili(t *testing.T) {
	req, err := http.NewRequest(http.MethodGet, "https://example.com", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Range", "bytes=0-99")
	applyMediaHeaders(req, "https://upos-hz-mirrorakam.akamaized.net/video.m4s")

	if got := req.Header.Get("Referer"); got != "https://www.bilibili.com" {
		t.Fatalf("Referer = %q", got)
	}
	if got := req.Header.Get("Accept"); got != "*/*" {
		t.Fatalf("Accept = %q", got)
	}
	if got := req.Header.Get("Connection"); got != "close" {
		t.Fatalf("Connection = %q", got)
	}
}

func TestApplyMediaHeadersForNicoCookie(t *testing.T) {
	req, err := http.NewRequest(http.MethodGet, "https://example.com", nil)
	if err != nil {
		t.Fatal(err)
	}
	applyMediaHeaders(req, "https://delivery.domand.nicovideo.jp/seg.ts#cookie=domand_bid%3Dabc123")

	if got := req.Header.Get("Cookie"); got != "domand_bid=abc123" {
		t.Fatalf("Cookie = %q", got)
	}
}
