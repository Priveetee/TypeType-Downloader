package pipeline

import (
	"math"

	"typetype-downloader-go/internal/selector"
)

const (
	outputHeadroomTenths = uint64(11)
	localPeakTenths      = uint64(21)
)

func (r *Runner) reserveVideo(id string, selection *selector.Selection, duration int64) (func(), error) {
	bytes := videoReservationBytes(selection, duration)
	if bytes == 0 {
		return noRelease, nil
	}
	return r.disk.Reserve(id, bytes)
}

func (r *Runner) reserveAudio(id string, selection *selector.AudioSelection, duration int64) (func(), error) {
	bytes := audioReservationBytes(selection, duration)
	if bytes == 0 {
		return noRelease, nil
	}
	return r.disk.Reserve(id, bytes)
}

func videoReservationBytes(selection *selector.Selection, duration int64) uint64 {
	videoBytes := mediaBytes(selection.Video.ContentLength, selection.Video.Bitrate, duration)
	audioBytes := mediaBytes(selection.Audio.ContentLength, selection.Audio.Bitrate, duration)
	total := knownSum(videoBytes, audioBytes)
	if total == 0 {
		return 0
	}
	multiplier := outputHeadroomTenths
	if selection.Video.DeliveryMethod == "sabr" ||
		selection.Video.ContentLength > 0 && selection.Audio.ContentLength > 0 && !usesRemoteMux(selection) {
		multiplier = localPeakTenths
	}
	return scaledBytes(total, multiplier)
}

func audioReservationBytes(selection *selector.AudioSelection, duration int64) uint64 {
	bytes := mediaBytes(selection.Audio.ContentLength, selection.Audio.Bitrate, duration)
	if bytes == 0 {
		return 0
	}
	multiplier := outputHeadroomTenths
	if selection.Audio.DeliveryMethod == "sabr" {
		multiplier = localPeakTenths
	}
	return scaledBytes(bytes, multiplier)
}

func mediaBytes(contentLength int64, bitrate *int, duration int64) uint64 {
	if contentLength > 0 {
		return uint64(contentLength)
	}
	if bitrate == nil || *bitrate <= 0 || duration <= 0 {
		return 0
	}
	rate := uint64(*bitrate)
	if rate > math.MaxUint64/uint64(duration) {
		return math.MaxUint64
	}
	return rate * uint64(duration) / 8
}

func knownSum(left uint64, right uint64) uint64 {
	if left > math.MaxUint64-right {
		return math.MaxUint64
	}
	return left + right
}

func scaledBytes(bytes uint64, tenths uint64) uint64 {
	if bytes > math.MaxUint64/tenths {
		return math.MaxUint64
	}
	return bytes * tenths / 10
}

func noRelease() {}
