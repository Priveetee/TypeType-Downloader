package pipeline

import (
	"math"
	"testing"

	"typetype-downloader-go/internal/selector"
	"typetype-downloader-go/internal/typetype"
)

func TestKnownSumUsesAvailableInputs(t *testing.T) {
	if got := knownSum(100, 50); got != 150 {
		t.Fatalf("sum = %d", got)
	}
	if got := knownSum(100, 0); got != 100 {
		t.Fatalf("partial sum = %d", got)
	}
}

func TestMediaBytesPrefersContentLengthAndEstimatesFallback(t *testing.T) {
	bitrate := 128_000
	if got := mediaBytes(42, &bitrate, 36_000); got != 42 {
		t.Fatalf("content bytes = %d", got)
	}
	if got := mediaBytes(0, &bitrate, 36_000); got != 576_000_000 {
		t.Fatalf("estimated bytes = %d", got)
	}
}

func TestCapacityEstimateSaturates(t *testing.T) {
	if got := scaledBytes(100, 21); got != 210 {
		t.Fatalf("scaled bytes = %d", got)
	}
	if got := scaledBytes(math.MaxUint64, 21); got != math.MaxUint64 {
		t.Fatalf("saturated bytes = %d", got)
	}
}

func TestVideoReservationCoversLocalAssemblyPeak(t *testing.T) {
	selection := &selector.Selection{
		Video: typetype.VideoStreamItem{URL: "video", ContentLength: 100},
		Audio: typetype.AudioStreamItem{URL: "audio", ContentLength: 50},
	}
	if got := videoReservationBytes(selection, 0); got != 315 {
		t.Fatalf("local reservation = %d", got)
	}
	selection.Video.URL = "https://example.test/video.m3u8"
	if got := videoReservationBytes(selection, 0); got != 165 {
		t.Fatalf("remote reservation = %d", got)
	}
	selection.Video.DeliveryMethod = "sabr"
	if got := videoReservationBytes(selection, 0); got != 315 {
		t.Fatalf("SABR reservation = %d", got)
	}
}

func TestAudioReservationCoversSABRAssemblyPeak(t *testing.T) {
	selection := &selector.AudioSelection{
		Audio: typetype.AudioStreamItem{ContentLength: 100},
	}
	if got := audioReservationBytes(selection, 0); got != 110 {
		t.Fatalf("direct reservation = %d", got)
	}
	selection.Audio.DeliveryMethod = "sabr"
	if got := audioReservationBytes(selection, 0); got != 210 {
		t.Fatalf("SABR reservation = %d", got)
	}
}
