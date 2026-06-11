package typetype

import "testing"

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

func TestNormalizeStreamURLsResolvesRelativeProxyURLs(t *testing.T) {
	client := NewClient("http://typetype-server:8080")
	stream := &StreamResponse{
		VideoOnlyStreams: []VideoStreamItem{{URL: "proxy?url=https%3A%2F%2Fexample.com%2Fvideo.mp4"}},
		AudioStreams: []AudioStreamItem{
			{URL: "/proxy?url=https%3A%2F%2Fexample.com%2Faudio.m4a"},
			{URL: "nicovideo?url=https%3A%2F%2Fexample.com%2Fseg.ts"},
		},
	}

	client.normalizeStreamURLs(stream)

	if got, want := stream.VideoOnlyStreams[0].URL, "http://typetype-server:8080/proxy?url=https%3A%2F%2Fexample.com%2Fvideo.mp4"; got != want {
		t.Fatalf("video URL = %q, want %q", got, want)
	}
	if got, want := stream.AudioStreams[0].URL, "http://typetype-server:8080/proxy?url=https%3A%2F%2Fexample.com%2Faudio.m4a"; got != want {
		t.Fatalf("audio URL = %q, want %q", got, want)
	}
	if got, want := stream.AudioStreams[1].URL, "http://typetype-server:8080/proxy/nicovideo?url=https%3A%2F%2Fexample.com%2Fseg.ts"; got != want {
		t.Fatalf("nico URL = %q, want %q", got, want)
	}
}

func TestNormalizeStreamURLsKeepsAbsoluteMediaURLsDirect(t *testing.T) {
	client := NewClient("http://typetype-server:8080")
	raw := "https://upos-hz-mirrorakam.akamaized.net/video.m4s?deadline=10000&upsig=x"
	stream := &StreamResponse{VideoOnlyStreams: []VideoStreamItem{{URL: raw}}}

	client.normalizeStreamURLs(stream)

	got := stream.VideoOnlyStreams[0].URL
	if got != raw {
		t.Fatalf("media URL = %q, want %q", got, raw)
	}
}
