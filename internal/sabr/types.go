package sabr

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
