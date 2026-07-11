package pipeline

import (
	"strconv"
	"strings"

	"typetype-downloader-go/internal/job"
	"typetype-downloader-go/internal/selector"
)

func selectorOptions(options job.Options) selector.Options {
	container := strings.TrimSpace(options.Container)
	if container == "" {
		container = strings.TrimSpace(options.Format)
	}
	if container == "" {
		container = "mp4"
	}
	maxHeight := 1080
	if options.Height != nil && *options.Height > 0 {
		maxHeight = *options.Height
	}
	return selector.Options{
		Container:        container,
		MaxHeight:        maxHeight,
		VideoCodecPrefix: codecPrefix(options.VideoCodec),
		AudioCodecPrefix: codecPrefix(options.AudioCodec),
		VideoItag:        parseItag(options.VideoItag),
		AudioItag:        parseItag(options.AudioItag),
	}
}

func codecPrefix(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "h264", "x264":
		return "avc1"
	case "h265", "hevc", "x265":
		return "hev1"
	case "av1":
		return "av01"
	case "aac":
		return "mp4a"
	default:
		return strings.ToLower(strings.TrimSpace(value))
	}
}

func isAudioOnly(options job.Options) bool {
	mode := strings.TrimSpace(strings.ToLower(options.Mode))
	return mode == "audio"
}

func parseItag(value string) int {
	itag, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil {
		return 0
	}
	return itag
}
