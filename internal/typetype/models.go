package typetype

type StreamResponse struct {
	ID                      string            `json:"id"`
	Title                   string            `json:"title"`
	Duration                int64             `json:"duration"`
	VideoStreams            []VideoStreamItem `json:"videoStreams"`
	VideoOnlyStreams        []VideoStreamItem `json:"videoOnlyStreams"`
	AudioStreams            []AudioStreamItem `json:"audioStreams"`
	OriginalAudioTrackID    *string           `json:"originalAudioTrackId"`
	PreferredDefaultAudioID *string           `json:"preferredDefaultAudioTrackId"`
	HLSURL                  string            `json:"hlsUrl"`
	DashMPDURL              string            `json:"dashMpdUrl"`
}

type VideoStreamItem struct {
	URL            string  `json:"url"`
	MimeType       string  `json:"mimeType"`
	Format         string  `json:"format"`
	Resolution     string  `json:"resolution"`
	Bitrate        *int    `json:"bitrate"`
	Codec          *string `json:"codec"`
	IsVideoOnly    bool    `json:"isVideoOnly"`
	Itag           int     `json:"itag"`
	Width          int     `json:"width"`
	Height         int     `json:"height"`
	FPS            int     `json:"fps"`
	ContentLength  int64   `json:"contentLength"`
	InitStart      int64   `json:"initStart"`
	InitEnd        int64   `json:"initEnd"`
	IndexStart     int64   `json:"indexStart"`
	IndexEnd       int64   `json:"indexEnd"`
	DeliveryMethod string  `json:"deliveryMethod"`
	ManifestURL    string  `json:"manifestUrl"`
	SABRSessionURL string  `json:"sabrSessionUrl"`
}

type AudioStreamItem struct {
	URL            string  `json:"url"`
	MimeType       string  `json:"mimeType"`
	Format         string  `json:"format"`
	Bitrate        *int    `json:"bitrate"`
	Codec          *string `json:"codec"`
	Quality        *string `json:"quality"`
	Itag           int     `json:"itag"`
	ContentLength  int64   `json:"contentLength"`
	InitStart      int64   `json:"initStart"`
	InitEnd        int64   `json:"initEnd"`
	IndexStart     int64   `json:"indexStart"`
	IndexEnd       int64   `json:"indexEnd"`
	AudioTrackID   *string `json:"audioTrackId"`
	AudioTrackName *string `json:"audioTrackName"`
	AudioLocale    *string `json:"audioLocale"`
	IsOriginal     bool    `json:"isOriginal"`
	DeliveryMethod string  `json:"deliveryMethod"`
	ManifestURL    string  `json:"manifestUrl"`
	SABRSessionURL string  `json:"sabrSessionUrl"`
}
