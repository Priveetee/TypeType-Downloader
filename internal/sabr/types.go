package sabr

import "os"

type Options struct {
	ManifestURL   string
	Authorization string
	VideoItag     int
	AudioItag     int
	AudioTrackID  string
	AudioOnly     bool
	Workers       int
	WorkDir       string
	VideoPath     string
	AudioPath     string
}

type ProgressFunc func(downloadedBytes int64)

type filePlan struct {
	URL  string
	Path string
}

type trackPlan struct {
	Parts  []string
	Target string
}

type streamTrack struct {
	kind         string
	itag         int
	path         string
	output       *os.File
	nextSequence int
	initialized  bool
	mediaWritten bool
}
