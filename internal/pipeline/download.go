package pipeline

import (
	"context"
	"sync"
	"time"

	"typetype-downloader-go/internal/artifact"
	"typetype-downloader-go/internal/downloader"
	"typetype-downloader-go/internal/job"
	"typetype-downloader-go/internal/selector"
)

func (r *Runner) download(ctx context.Context, id string, selection *selector.Selection, paths artifact.Paths) (int64, error) {
	started := time.Now()
	total := selection.Video.ContentLength + selection.Audio.ContentLength
	progress := combinedProgress(func(downloaded int64, speed int64) {
		r.store.Progress(id, job.Progress{Stage: "download", DownloadedBytes: downloaded, TotalBytes: total, SpeedBytesPerSecond: speed})
	})
	options := downloader.Options{
		ChunkSize:     r.cfg.ChunkSize,
		Workers:       r.cfg.DownloadWorkers,
		Retries:       4,
		BufferSize:    256 * 1024,
		RangeMode:     r.cfg.RangeMode,
		ProgressBytes: 16 << 20,
	}
	err := runParallel(
		func() error {
			return downloader.DownloadFile(ctx, r.http, downloader.Source{Name: "video", URL: selection.Video.URL, Size: selection.Video.ContentLength}, paths.Video, options, progress)
		},
		func() error {
			return downloader.DownloadFile(ctx, r.http, downloader.Source{Name: "audio", URL: selection.Audio.URL, Size: selection.Audio.ContentLength}, paths.Audio, options, progress)
		},
	)
	return time.Since(started).Milliseconds(), err
}

func (r *Runner) downloadAudio(ctx context.Context, id string, selection *selector.AudioSelection, paths artifact.Paths) (int64, error) {
	started := time.Now()
	total := selection.Audio.ContentLength
	progress := combinedProgress(func(downloaded int64, speed int64) {
		r.store.Progress(id, job.Progress{Stage: "download", DownloadedBytes: downloaded, TotalBytes: total, SpeedBytesPerSecond: speed})
	})
	options := downloader.Options{
		ChunkSize:     r.cfg.ChunkSize,
		Workers:       r.cfg.DownloadWorkers,
		Retries:       4,
		BufferSize:    256 * 1024,
		RangeMode:     r.cfg.RangeMode,
		ProgressBytes: 4 << 20,
	}
	err := downloader.DownloadFile(ctx, r.http, downloader.Source{Name: "audio", URL: selection.Audio.URL, Size: selection.Audio.ContentLength}, paths.Output, options, progress)
	return time.Since(started).Milliseconds(), err
}

func combinedProgress(update func(downloaded int64, speed int64)) downloader.ProgressFunc {
	var mu sync.Mutex
	downloadedByName := map[string]int64{}
	speedByName := map[string]int64{}
	return func(progress downloader.Progress) {
		mu.Lock()
		defer mu.Unlock()
		downloadedByName[progress.Name] = progress.Downloaded
		speedByName[progress.Name] = int64(progress.Speed)
		var downloaded int64
		var speed int64
		for _, value := range downloadedByName {
			downloaded += value
		}
		for _, value := range speedByName {
			speed += value
		}
		update(downloaded, speed)
	}
}
