package job

import "time"

type Status string

const (
	StatusQueued  Status = "queued"
	StatusRunning Status = "running"
	StatusDone    Status = "done"
	StatusFailed  Status = "failed"
)

type CreateRequest struct {
	URL     string  `json:"url"`
	Options Options `json:"options"`
}

type CreateResponse struct {
	ID     string `json:"id"`
	Cached bool   `json:"cached"`
}

type Options struct {
	Mode                   string   `json:"mode"`
	Quality                string   `json:"quality"`
	Format                 string   `json:"format"`
	Container              string   `json:"container"`
	AudioPassthrough       bool     `json:"audioPassthrough"`
	VideoItag              string   `json:"videoItag"`
	AudioItag              string   `json:"audioItag"`
	Height                 *int     `json:"height"`
	FPS                    *int     `json:"fps"`
	VideoCodec             string   `json:"videoCodec"`
	AudioCodec             string   `json:"audioCodec"`
	Bitrate                *int     `json:"bitrate"`
	AllowQualityFallback   bool     `json:"allowQualityFallback"`
	SponsorBlock           bool     `json:"sponsorBlock"`
	SponsorBlockCategories []string `json:"sponsorBlockCategories"`
	ThumbnailOnly          bool     `json:"thumbnailOnly"`
	Subtitles              any      `json:"subtitles"`
}

type Response struct {
	ID                  string          `json:"id"`
	URL                 string          `json:"url"`
	Status              Status          `json:"status"`
	DurationMs          int64           `json:"durationMs"`
	Title               string          `json:"title"`
	Error               *string         `json:"error,omitempty"`
	ErrorCode           *string         `json:"errorCode,omitempty"`
	ArtifactURL         *string         `json:"artifactUrl,omitempty"`
	ArtifactExpiresAt   *string         `json:"artifactExpiresAt,omitempty"`
	Resolved            *ResolvedOutput `json:"resolved,omitempty"`
	ProgressPercent     *int            `json:"progressPercent,omitempty"`
	DownloadedBytes     *int64          `json:"downloadedBytes,omitempty"`
	TotalBytes          *int64          `json:"totalBytes,omitempty"`
	ETASeconds          *int64          `json:"etaSeconds,omitempty"`
	SpeedBytesPerSecond *int64          `json:"speedBytesPerSecond,omitempty"`
	Stage               *string         `json:"stage,omitempty"`
	QueuedAt            *string         `json:"queuedAt,omitempty"`
	StartedAt           *string         `json:"startedAt,omitempty"`
	FinishedAt          *string         `json:"finishedAt,omitempty"`
	QueueWaitMs         *int64          `json:"queueWaitMs,omitempty"`
	RunTimeMs           *int64          `json:"runTimeMs,omitempty"`
	TokenFetchMs        *int64          `json:"tokenFetchMs,omitempty"`
	YtdlpMs             *int64          `json:"ytdlpMs,omitempty"`
	UploadMs            *int64          `json:"uploadMs,omitempty"`
	DownloadMs          *int64          `json:"downloadMs,omitempty"`
	MuxMs               *int64          `json:"muxMs,omitempty"`
	TotalMs             *int64          `json:"totalMs,omitempty"`
}

type ResolvedOutput struct {
	VideoItag  string `json:"videoItag,omitempty"`
	AudioItag  string `json:"audioItag,omitempty"`
	Height     int    `json:"height,omitempty"`
	FPS        int    `json:"fps,omitempty"`
	VideoCodec string `json:"videoCodec,omitempty"`
	AudioCodec string `json:"audioCodec,omitempty"`
	Container  string `json:"container,omitempty"`
	FileName   string `json:"fileName,omitempty"`
}

type Progress struct {
	Stage               string
	DownloadedBytes     int64
	TotalBytes          int64
	SpeedBytesPerSecond int64
}

type Record struct {
	ID            string
	CacheKey      string
	URL           string
	Authorization string
	Options       Options
	Status        Status
	Title         string
	Error         *string
	ErrorCode     *string
	Artifact      string
	ExpiresAt     *time.Time
	Storage       string
	Resolved      *ResolvedOutput
	Progress      Progress
	QueuedAt      time.Time
	StartedAt     *time.Time
	FinishedAt    *time.Time
	DownloadMs    *int64
	MuxMs         *int64
	TotalMs       *int64
	Cancel        func()
}
