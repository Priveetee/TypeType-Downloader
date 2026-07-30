package pipeline

import (
	"context"
	"os"
	"path/filepath"
	"time"

	"typetype-downloader-go/internal/artifact"
	"typetype-downloader-go/internal/job"
	"typetype-downloader-go/internal/selector"
)

func (r *Runner) runSABR(ctx context.Context, id string, record *job.Record, title string, selection *selector.Selection) error {
	paths := artifact.Build(r.cfg.DataDir, id, title, selection.Container)
	r.store.Resolve(id, title, resolvedOutput(selection, paths.Name))
	return r.runSABRArtifact(ctx, id, paths, func() (artifactTimings, error) {
		downloadMs, err := r.downloadSABR(ctx, id, record, selection, paths)
		if err != nil {
			return artifactTimings{downloadMs: downloadMs}, err
		}
		totalBytes := selection.Video.ContentLength + selection.Audio.ContentLength
		r.store.Progress(id, job.Progress{Stage: "mux", DownloadedBytes: totalBytes, TotalBytes: totalBytes})
		muxStarted := time.Now()
		err = retry(ctx, 2, "mux", func() error {
			_ = os.Remove(paths.Output)
			return merge(ctx, r.cfg.Muxer, paths.Video, paths.Audio, paths.Output)
		})
		return artifactTimings{
			downloadMs: downloadMs,
			muxMs:      time.Since(muxStarted).Milliseconds(),
		}, err
	})
}

func (r *Runner) runSABRAudio(ctx context.Context, id string, record *job.Record, title string, selection *selector.AudioSelection) error {
	paths := artifact.Build(r.cfg.DataDir, id, title, selection.Container)
	r.store.Resolve(id, title, audioResolvedOutput(selection, paths.Name))
	return r.runSABRArtifact(ctx, id, paths, func() (artifactTimings, error) {
		downloadMs, err := r.downloadSABRAudio(ctx, id, record, selection, paths)
		return artifactTimings{downloadMs: downloadMs}, err
	})
}

func (r *Runner) runSABRArtifact(
	ctx context.Context,
	id string,
	paths artifact.Paths,
	process func() (artifactTimings, error),
) error {
	preserveOutput := false
	defer func() { cleanupWork(paths, preserveOutput) }()
	if err := os.MkdirAll(paths.WorkDir, 0o755); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(paths.Output), 0o755); err != nil {
		return err
	}
	timings, err := process()
	if err != nil {
		return err
	}
	var saved artifact.Saved
	if err := retry(ctx, 3, "artifact upload", func() error {
		var saveErr error
		saved, saveErr = r.storage.Save(ctx, paths.Output, paths.Key)
		return saveErr
	}); err != nil {
		return err
	}
	var expires *time.Time
	if !saved.Expires.IsZero() {
		expires = &saved.Expires
	}
	r.store.Done(id, saved.Location, saved.Backend, expires, timings.downloadMs, timings.muxMs)
	preserveOutput = saved.Backend == "local"
	return nil
}

type artifactTimings struct {
	downloadMs int64
	muxMs      int64
}
