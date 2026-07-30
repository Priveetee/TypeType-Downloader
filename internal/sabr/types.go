package sabr

import (
	"os"
	"time"
)

type Options struct {
	ManifestURL   string
	Authorization string
	VideoItag     int
	AudioItag     int
	AudioTrackID  string
	AudioOnly     bool
	VideoPath     string
	AudioPath     string
	ExpectedBytes int64
	Parts         int
	IdleTimeout   time.Duration
}

type ProgressFunc func(downloadedBytes int64)

type streamTrack struct {
	kind         string
	itag         int
	path         string
	output       *os.File
	nextSequence int
	initialized  bool
	mediaWritten bool
}
