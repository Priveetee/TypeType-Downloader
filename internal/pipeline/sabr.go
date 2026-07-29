package pipeline

import (
	"context"
	"time"

	"typetype-downloader-go/internal/artifact"
	"typetype-downloader-go/internal/job"
	"typetype-downloader-go/internal/sabr"
	"typetype-downloader-go/internal/selector"
)

func (r *Runner) downloadSABR(ctx context.Context, id string, record *job.Record, selection *selector.Selection, paths artifact.Paths) (int64, error) {
	started := time.Now()
	totalBytes := selection.Video.ContentLength + selection.Audio.ContentLength
	r.store.Progress(id, job.Progress{Stage: "download", TotalBytes: totalBytes})
	trackID := ""
	if selection.Audio.AudioTrackID != nil {
		trackID = *selection.Audio.AudioTrackID
	}
	err := sabr.Download(ctx, r.http, sabr.Options{
		ManifestURL:   r.streams.MediaURL(selection.Video.ManifestURL),
		Authorization: record.Authorization,
		VideoItag:     selection.Video.Itag,
		AudioItag:     selection.Audio.Itag,
		AudioTrackID:  trackID,
		VideoPath:     paths.Video,
		AudioPath:     paths.Audio,
		ExpectedBytes: totalBytes,
	}, r.sabrProgress(id, started, totalBytes))
	return time.Since(started).Milliseconds(), err
}

func (r *Runner) downloadSABRAudio(ctx context.Context, id string, record *job.Record, selection *selector.AudioSelection, paths artifact.Paths) (int64, error) {
	started := time.Now()
	totalBytes := selection.Audio.ContentLength
	r.store.Progress(id, job.Progress{Stage: "download", TotalBytes: totalBytes})
	trackID := ""
	if selection.Audio.AudioTrackID != nil {
		trackID = *selection.Audio.AudioTrackID
	}
	err := sabr.Download(ctx, r.http, sabr.Options{
		ManifestURL:   r.streams.MediaURL(selection.Audio.ManifestURL),
		Authorization: record.Authorization,
		AudioItag:     selection.Audio.Itag,
		AudioTrackID:  trackID,
		AudioOnly:     true,
		AudioPath:     paths.Output,
		ExpectedBytes: totalBytes,
	}, r.sabrProgress(id, started, totalBytes))
	return time.Since(started).Milliseconds(), err
}

func (r *Runner) sabrProgress(id string, started time.Time, totalBytes int64) sabr.ProgressFunc {
	return func(downloadedBytes int64) {
		if totalBytes > 0 && downloadedBytes > totalBytes {
			downloadedBytes = totalBytes
		}
		elapsed := time.Since(started).Seconds()
		speed := int64(0)
		if elapsed > 0 {
			speed = int64(float64(downloadedBytes) / elapsed)
		}
		r.store.Progress(id, job.Progress{
			Stage:               "download",
			DownloadedBytes:     downloadedBytes,
			TotalBytes:          totalBytes,
			SpeedBytesPerSecond: speed,
		})
	}
}
