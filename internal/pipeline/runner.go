package pipeline

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"typetype-downloader-go/internal/artifact"
	"typetype-downloader-go/internal/config"
	"typetype-downloader-go/internal/job"
	"typetype-downloader-go/internal/selector"
	"typetype-downloader-go/internal/typetype"
)

type Runner struct {
	cfg     config.Config
	store   *job.Store
	streams *typetype.Client
	storage artifact.Store
	http    *http.Client
	queue   chan string
}

func NewRunner(cfg config.Config, store *job.Store, storage artifact.Store) *Runner {
	return &Runner{
		cfg:     cfg,
		store:   store,
		streams: typetype.NewClient(cfg.TypeTypeAPIBase),
		storage: storage,
		http:    newHTTPClient(cfg.DownloadWorkers, cfg.HTTP2),
		queue:   make(chan string, cfg.MaxQueueSize),
	}
}

func (r *Runner) Start(ctx context.Context) {
	for range r.cfg.MaxWorkers {
		go r.worker(ctx)
	}
}

func (r *Runner) Enqueue(id string) error {
	select {
	case r.queue <- id:
		return nil
	default:
		return fmt.Errorf("job queue is full")
	}
}

func (r *Runner) EnqueueBlocking(ctx context.Context, id string) error {
	select {
	case r.queue <- id:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (r *Runner) worker(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case id := <-r.queue:
			r.process(ctx, id)
		}
	}
}

func (r *Runner) process(parent context.Context, id string) {
	record, ok := r.store.Get(id)
	if !ok || record.Status != job.StatusQueued {
		return
	}
	ctx, cancel := context.WithCancel(parent)
	r.store.Start(id, cancel)
	defer cancel()
	if err := r.run(ctx, id, record); err != nil {
		code := "download_failed"
		if errors.Is(err, context.Canceled) {
			code = "cancelled"
		}
		r.store.Fail(id, code, err)
		slog.Warn("job failed", "id", id, "error", err)
	}
}

func (r *Runner) run(ctx context.Context, id string, record *job.Record) error {
	started := time.Now()
	var stream *typetype.StreamResponse
	if err := retry(ctx, 3, "stream extraction", func() error {
		var fetchErr error
		stream, fetchErr = r.streams.FetchStream(ctx, record.URL, record.Authorization)
		return fetchErr
	}); err != nil {
		return err
	}
	if isAudioOnly(record.Options) {
		return r.runAudioOnly(ctx, id, record, stream, started)
	}
	selection, err := selector.SelectMP4WithOptions(stream, selectorOptions(record.Options))
	if err != nil {
		return err
	}
	if selection.Video.DeliveryMethod == "sabr" {
		return r.runSABR(ctx, id, record, stream.Title, selection)
	}
	if selection.Video.ContentLength <= 0 || selection.Audio.ContentLength <= 0 {
		return r.runRemote(ctx, id, stream.Title, selection)
	}
	paths := artifact.Build(r.cfg.DataDir, id, stream.Title, selection.Container)
	preserveOutput := false
	defer func() { cleanupWork(paths, preserveOutput) }()
	resolved := resolvedOutput(selection, paths.Name)
	r.store.Resolve(id, stream.Title, resolved)
	if err := os.MkdirAll(paths.WorkDir, 0o755); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(paths.Output), 0o755); err != nil {
		return err
	}
	downloadMs, err := r.download(ctx, id, selection, paths)
	if err != nil {
		return err
	}
	muxStarted := time.Now()
	totalBytes := selection.Video.ContentLength + selection.Audio.ContentLength
	r.store.Progress(id, job.Progress{Stage: "mux", DownloadedBytes: totalBytes, TotalBytes: totalBytes})
	if err := retry(ctx, 2, "mux", func() error {
		_ = os.Remove(paths.Output)
		return merge(ctx, r.cfg.Muxer, paths.Video, paths.Audio, paths.Output)
	}); err != nil {
		return err
	}
	muxMs := time.Since(muxStarted).Milliseconds()
	var saved artifact.Saved
	if err := retry(ctx, 3, "artifact upload", func() error {
		var saveErr error
		saved, saveErr = r.storage.Save(ctx, paths.Output, paths.Key)
		return saveErr
	}); err != nil {
		return err
	}
	expiresAt := saved.Expires
	var expires *time.Time
	if !expiresAt.IsZero() {
		expires = &expiresAt
	}
	r.store.Done(id, saved.Location, saved.Backend, expires, downloadMs, muxMs)
	preserveOutput = saved.Backend == "local"
	slog.Info("job completed", "id", id, "ms", time.Since(started).Milliseconds())
	return nil
}

func (r *Runner) runAudioOnly(ctx context.Context, id string, record *job.Record, stream *typetype.StreamResponse, started time.Time) error {
	selection, err := selector.SelectAudioOnly(stream, selectorOptions(record.Options))
	if err != nil {
		return err
	}
	if selection.Audio.DeliveryMethod == "sabr" {
		return r.runSABRAudio(ctx, id, record, stream.Title, selection)
	}
	paths := artifact.Build(r.cfg.DataDir, id, stream.Title, selection.Container)
	preserveOutput := false
	defer func() { cleanupWork(paths, preserveOutput) }()
	resolved := audioResolvedOutput(selection, paths.Name)
	r.store.Resolve(id, stream.Title, resolved)
	if err := os.MkdirAll(paths.WorkDir, 0o755); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(paths.Output), 0o755); err != nil {
		return err
	}
	downloadMs, err := r.downloadAudio(ctx, id, selection, paths)
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
	expiresAt := saved.Expires
	var expires *time.Time
	if !expiresAt.IsZero() {
		expires = &expiresAt
	}
	r.store.Done(id, saved.Location, saved.Backend, expires, downloadMs, 0)
	preserveOutput = saved.Backend == "local"
	slog.Info("audio job completed", "id", id, "ms", time.Since(started).Milliseconds())
	return nil
}
