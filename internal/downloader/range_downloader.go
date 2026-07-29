package downloader

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"
)

type Source struct {
	Name string
	URL  string
	Size int64
}

type Options struct {
	ChunkSize     int64
	Workers       int
	Retries       int
	BufferSize    int
	RangeMode     string
	ProgressBytes int64
}

type Progress struct {
	Name       string
	Downloaded int64
	Total      int64
	Speed      float64
}

type ProgressFunc func(Progress)

func DownloadFile(
	ctx context.Context,
	client *http.Client,
	source Source,
	output string,
	options Options,
	progress ProgressFunc,
) error {
	if source.URL == "" {
		return fmt.Errorf("invalid source %s", source.Name)
	}
	if client == nil {
		client = http.DefaultClient
	}
	options = normalizedOptions(options)
	if source.Size <= 0 {
		size, err := probeSourceSize(ctx, client, source.URL, options.RangeMode)
		if err != nil {
			return fmt.Errorf("probe %s size: %w", source.Name, err)
		}
		source.Size = size
	}

	tmp := output + ".part"
	_ = os.Remove(tmp)
	file, err := os.OpenFile(tmp, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0o644)
	if err != nil {
		return err
	}
	committed := false
	defer func() {
		_ = file.Close()
		if !committed {
			_ = os.Remove(tmp)
		}
	}()
	if err := file.Truncate(source.Size); err != nil {
		return err
	}

	if err := downloadChunks(ctx, client, file, source, options, progress); err != nil {
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}
	if err := os.Rename(tmp, output); err != nil {
		return err
	}
	committed = true
	return nil
}

func normalizedOptions(options Options) Options {
	if options.ChunkSize <= 0 {
		options.ChunkSize = 10 << 20
	}
	if options.Workers <= 0 {
		options.Workers = 8
	}
	if options.Retries <= 0 {
		options.Retries = 4
	}
	if options.BufferSize <= 0 {
		options.BufferSize = defaultCopyBufferSize
	}
	if options.ProgressBytes <= 0 {
		options.ProgressBytes = 4 << 20
	}
	if options.RangeMode == "" {
		options.RangeMode = "header"
	}
	return options
}

func downloadChunks(
	ctx context.Context,
	client *http.Client,
	file *os.File,
	source Source,
	options Options,
	progress ProgressFunc,
) error {
	downloadCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	chunkCount := int((source.Size + options.ChunkSize - 1) / options.ChunkSize)
	workers := min(options.Workers, chunkCount)
	jobs := make(chan chunk, workers)
	errs := make(chan error, 1)
	tracker := newProgressTracker(source, options.ProgressBytes, progress)

	var wg sync.WaitGroup
	for range workers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			buffer := borrowCopyBuffer(options.BufferSize)
			defer releaseCopyBuffer(buffer)
			for part := range jobs {
				if err := downloadChunk(downloadCtx, client, file, source, part, options, buffer, tracker); err != nil {
					select {
					case errs <- err:
					default:
					}
					cancel()
					return
				}
			}
		}()
	}

	queueChunks(downloadCtx, jobs, source.Size, options.ChunkSize)
	close(jobs)
	wg.Wait()
	select {
	case err := <-errs:
		return err
	default:
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if got := tracker.downloaded.Load(); got != source.Size {
		return fmt.Errorf("downloaded %d bytes, expected %d", got, source.Size)
	}
	tracker.finish(time.Now())
	return nil
}

func queueChunks(ctx context.Context, jobs chan<- chunk, size int64, chunkSize int64) {
	for start := int64(0); start < size; start += chunkSize {
		end := min(start+chunkSize, size) - 1
		select {
		case <-ctx.Done():
			return
		case jobs <- chunk{start: start, end: end}:
		}
	}
}
