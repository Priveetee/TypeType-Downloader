package pipeline

import (
	"context"
	"time"

	"typetype-downloader-go/internal/artifact"
	"typetype-downloader-go/internal/ffmpeg"
	"typetype-downloader-go/internal/job"
	"typetype-downloader-go/internal/selector"
)

func (r *Runner) runRemote(ctx context.Context, id string, title string, selection *selector.Selection) error {
	paths := artifact.Build(r.cfg.DataDir, id, title, selection.Container)
	r.store.Resolve(id, title, resolvedOutput(selection, paths.Name))
	return r.runSABRArtifact(ctx, id, paths, func() (artifactTimings, error) {
		started := time.Now()
		r.store.Progress(id, job.Progress{Stage: "download"})
		err := ffmpeg.DownloadRemote(
			ctx,
			r.streams.ProxyMediaURL(selection.Video.URL),
			r.streams.ProxyMediaURL(selection.Audio.URL),
			paths.Output,
		)
		return artifactTimings{downloadMs: time.Since(started).Milliseconds()}, err
	})
}

func (r *Runner) runRemoteAudio(ctx context.Context, id string, title string, selection *selector.AudioSelection) error {
	paths := artifact.Build(r.cfg.DataDir, id, title, selection.Container)
	r.store.Resolve(id, title, audioResolvedOutput(selection, paths.Name))
	return r.runSABRArtifact(ctx, id, paths, func() (artifactTimings, error) {
		started := time.Now()
		r.store.Progress(id, job.Progress{Stage: "download"})
		err := ffmpeg.DownloadRemoteAudio(ctx, r.streams.ProxyMediaURL(selection.Audio.URL), paths.Output)
		return artifactTimings{downloadMs: time.Since(started).Milliseconds()}, err
	})
}
