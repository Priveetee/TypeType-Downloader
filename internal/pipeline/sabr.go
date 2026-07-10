package pipeline

import (
	"context"
	"time"

	"typetype-downloader-go/internal/ffmpeg"
	"typetype-downloader-go/internal/job"
	"typetype-downloader-go/internal/selector"
)

func (r *Runner) downloadSABR(ctx context.Context, id string, record *job.Record, selection *selector.Selection, output string) (int64, error) {
	started := time.Now()
	r.store.Progress(id, job.Progress{Stage: "download"})
	trackID := ""
	if selection.Audio.AudioTrackID != nil {
		trackID = *selection.Audio.AudioTrackID
	}
	err := ffmpeg.DownloadSABR(ctx, ffmpeg.SABROptions{
		ManifestURL:   r.streams.MediaURL(selection.Video.ManifestURL),
		Authorization: record.Authorization,
		VideoItag:     selection.Video.Itag,
		AudioItag:     selection.Audio.Itag,
		AudioTrackID:  trackID,
	}, output)
	return time.Since(started).Milliseconds(), err
}

func (r *Runner) downloadSABRAudio(ctx context.Context, id string, record *job.Record, selection *selector.AudioSelection, output string) (int64, error) {
	started := time.Now()
	r.store.Progress(id, job.Progress{Stage: "download"})
	trackID := ""
	if selection.Audio.AudioTrackID != nil {
		trackID = *selection.Audio.AudioTrackID
	}
	err := ffmpeg.DownloadSABR(ctx, ffmpeg.SABROptions{
		ManifestURL:   r.streams.MediaURL(selection.Audio.ManifestURL),
		Authorization: record.Authorization,
		AudioItag:     selection.Audio.Itag,
		AudioTrackID:  trackID,
		AudioOnly:     true,
	}, output)
	return time.Since(started).Milliseconds(), err
}
