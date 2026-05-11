package pipeline

import (
	"testing"

	"typetype-downloader-go/internal/job"
)

func TestSelectorOptionsUsesFormatWhenContainerMissing(t *testing.T) {
	height := 720
	options := selectorOptions(job.Options{
		Format:     "webm",
		Height:     &height,
		VideoItag:  "248",
		AudioItag:  "251",
		VideoCodec: "vp9",
		AudioCodec: "opus",
	})
	if options.Container != "webm" || options.MaxHeight != 720 {
		t.Fatalf("container/height = %s/%d, want webm/720", options.Container, options.MaxHeight)
	}
	if options.VideoItag != 248 || options.AudioItag != 251 {
		t.Fatalf("itags = %d+%d, want 248+251", options.VideoItag, options.AudioItag)
	}
	if options.VideoCodecPrefix != "vp9" || options.AudioCodecPrefix != "opus" {
		t.Fatalf("codecs = %s/%s, want vp9/opus", options.VideoCodecPrefix, options.AudioCodecPrefix)
	}
}

func TestIsAudioOnlyIsCaseInsensitive(t *testing.T) {
	if !isAudioOnly(job.Options{Mode: " AUDIO "}) {
		t.Fatal("expected AUDIO mode to be audio-only")
	}
	if isAudioOnly(job.Options{Mode: "video"}) {
		t.Fatal("expected video mode to not be audio-only")
	}
}
