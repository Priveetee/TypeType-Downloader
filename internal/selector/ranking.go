package selector

import "typetype-downloader-go/internal/typetype"

func playableVideo(stream typetype.VideoStreamItem) bool {
	return stream.URL != "" && stream.ContentLength > 0 || stream.DeliveryMethod == "sabr" && stream.ManifestURL != ""
}

func playableAudio(stream typetype.AudioStreamItem) bool {
	return stream.URL != "" && stream.ContentLength > 0 || stream.DeliveryMethod == "sabr" && stream.ManifestURL != ""
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
