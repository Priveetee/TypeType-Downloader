package selector

import (
	"fmt"
	"strings"

	"typetype-downloader-go/internal/typetype"
)

type Selection struct {
	Title     string
	Video     typetype.VideoStreamItem
	Audio     typetype.AudioStreamItem
	Container string
}

type AudioSelection struct {
	Title     string
	Audio     typetype.AudioStreamItem
	Container string
}

type Options struct {
	Container        string
	MaxHeight        int
	VideoCodecPrefix string
	AudioCodecPrefix string
	VideoItag        int
	AudioItag        int
}

func SelectMP4(stream *typetype.StreamResponse, maxHeight int) (*Selection, error) {
	return SelectMP4WithOptions(stream, Options{
		MaxHeight:        maxHeight,
		VideoCodecPrefix: "avc1",
		AudioCodecPrefix: "mp4a",
	})
}

func SelectMP4WithOptions(stream *typetype.StreamResponse, options Options) (*Selection, error) {
	if options.Container == "" {
		options.Container = "mp4"
	}
	if options.MaxHeight <= 0 {
		options.MaxHeight = 1080
	}
	if options.VideoCodecPrefix == "" {
		if options.Container == "webm" {
			options.VideoCodecPrefix = "vp9"
		} else {
			options.VideoCodecPrefix = "avc1"
		}
	}
	if options.AudioCodecPrefix == "" {
		if options.Container == "webm" {
			options.AudioCodecPrefix = "opus"
		} else {
			options.AudioCodecPrefix = "mp4a"
		}
	}
	video, ok := selectVideo(stream.VideoOnlyStreams, options)
	if !ok {
		return nil, fmt.Errorf("no compatible mp4 video-only stream for codec %s", options.VideoCodecPrefix)
	}
	audio, ok := selectAudio(stream.AudioStreams, stream.PreferredDefaultAudioID, options)
	if !ok {
		return nil, fmt.Errorf("no compatible m4a audio stream for codec %s", options.AudioCodecPrefix)
	}
	return &Selection{
		Title:     stream.Title,
		Video:     video,
		Audio:     audio,
		Container: options.Container,
	}, nil
}

func SelectAudioOnly(stream *typetype.StreamResponse, options Options) (*AudioSelection, error) {
	container := audioOutputContainer(options.Container)
	matchContainer := container
	if matchContainer == "m4a" {
		matchContainer = "mp4"
	}
	codec := options.AudioCodecPrefix
	if codec == "" {
		if matchContainer == "webm" {
			codec = "opus"
		} else {
			codec = "mp4a"
		}
	}
	options.Container = matchContainer
	options.AudioCodecPrefix = codec
	audio, ok := selectAudio(stream.AudioStreams, stream.PreferredDefaultAudioID, options)
	if !ok {
		return nil, fmt.Errorf("no compatible audio stream for codec %s", codec)
	}
	return &AudioSelection{Title: stream.Title, Audio: audio, Container: container}, nil
}

func audioOutputContainer(container string) string {
	container = strings.TrimSpace(strings.ToLower(container))
	if container == "webm" || container == "opus" {
		return "webm"
	}
	return "m4a"
}

func selectVideo(streams []typetype.VideoStreamItem, options Options) (typetype.VideoStreamItem, bool) {
	var best typetype.VideoStreamItem
	found := false
	for _, stream := range streams {
		codec := stringValue(stream.Codec)
		if options.VideoItag > 0 && stream.Itag != options.VideoItag {
			continue
		}
		if stream.URL == "" || stream.ContentLength <= 0 || !strings.Contains(stream.MimeType, options.Container) || !strings.HasPrefix(codec, options.VideoCodecPrefix) {
			continue
		}
		if options.MaxHeight > 0 && stream.Height > options.MaxHeight {
			continue
		}
		if !found || stream.Height > best.Height || stream.Height == best.Height && streamBitrate(stream.Bitrate) > streamBitrate(best.Bitrate) {
			best = stream
			found = true
		}
	}
	return best, found
}

func selectAudio(streams []typetype.AudioStreamItem, preferredTrackID *string, options Options) (typetype.AudioStreamItem, bool) {
	var best typetype.AudioStreamItem
	found := false
	for _, stream := range streams {
		codec := stringValue(stream.Codec)
		if options.AudioItag > 0 && stream.Itag != options.AudioItag {
			continue
		}
		if stream.URL == "" || stream.ContentLength <= 0 || !strings.Contains(stream.MimeType, options.Container) || !strings.HasPrefix(codec, options.AudioCodecPrefix) {
			continue
		}
		if !found || audioRank(stream, preferredTrackID) > audioRank(best, preferredTrackID) {
			best = stream
			found = true
		}
	}
	return best, found
}

func audioRank(stream typetype.AudioStreamItem, preferredTrackID *string) int {
	rank := streamBitrate(stream.Bitrate)
	if preferredTrackID != nil && stream.AudioTrackID != nil && *preferredTrackID == *stream.AudioTrackID {
		rank += 1_000_000
	}
	if stream.IsOriginal {
		rank += 500_000
	}
	return rank
}

func streamBitrate(value *int) int {
	if value == nil {
		return 0
	}
	return *value
}

func stringValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}
