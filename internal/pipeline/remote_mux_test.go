package pipeline

import (
	"testing"

	"typetype-downloader-go/internal/selector"
	"typetype-downloader-go/internal/typetype"
)

func TestUsesRemoteMuxForHLSSelection(t *testing.T) {
	selection := &selector.Selection{
		Video: typetype.VideoStreamItem{URL: "https://media.example/video.M3U8?token=value"},
		Audio: typetype.AudioStreamItem{URL: "https://media.example/audio.m4a"},
	}
	if !usesRemoteMux(selection) {
		t.Fatal("expected HLS selection to use remote mux")
	}
}

func TestUsesRemoteMuxSkipsProgressiveSelection(t *testing.T) {
	selection := &selector.Selection{
		Video: typetype.VideoStreamItem{URL: "https://media.example/video.mp4"},
		Audio: typetype.AudioStreamItem{URL: "https://media.example/audio.m4a"},
	}
	if usesRemoteMux(selection) {
		t.Fatal("expected progressive selection to skip remote mux")
	}
}

func TestMediaPathStripsFragment(t *testing.T) {
	got, err := mediaPath("https://delivery.domand.nicovideo.jp/video.m3u8#cookie=domand_bid%3Dabc")
	if err != nil {
		t.Fatal(err)
	}
	if want := "https://delivery.domand.nicovideo.jp/video.m3u8"; got != want {
		t.Fatalf("mediaPath() = %q, want %q", got, want)
	}
}
