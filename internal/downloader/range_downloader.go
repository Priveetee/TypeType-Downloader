package downloader

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"sync"
	"sync/atomic"
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

func DownloadFile(ctx context.Context, client *http.Client, source Source, output string, options Options, progress ProgressFunc) error {
	if source.URL == "" {
		return fmt.Errorf("invalid source %s", source.Name)
	}
	if client == nil {
		client = http.DefaultClient
	}
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
		options.BufferSize = 256 * 1024
	}
	if options.ProgressBytes <= 0 {
		options.ProgressBytes = 4 << 20
	}
	if options.RangeMode == "" {
		options.RangeMode = "header"
	}
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
	defer file.Close()
	if err := file.Truncate(source.Size); err != nil {
		return err
	}

	downloadCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	jobs := make(chan chunk)
	errs := make(chan error, 1)
	var downloaded atomic.Int64
	started := time.Now()

	var wg sync.WaitGroup
	for range options.Workers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for job := range jobs {
				if err := downloadChunk(downloadCtx, client, file, source, job, options, &downloaded, started, progress); err != nil {
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

sendLoop:
	for start := int64(0); start < source.Size; start += options.ChunkSize {
		end := start + options.ChunkSize - 1
		if end >= source.Size {
			end = source.Size - 1
		}
		select {
		case <-downloadCtx.Done():
			break sendLoop
		case jobs <- chunk{start: start, end: end}:
		}
	}
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
	if got := downloaded.Load(); got != source.Size {
		return fmt.Errorf("downloaded %d bytes, expected %d", got, source.Size)
	}
	if err := file.Close(); err != nil {
		return err
	}
	return os.Rename(tmp, output)
}

type chunk struct {
	start int64
	end   int64
}

func downloadChunk(ctx context.Context, client *http.Client, file *os.File, source Source, part chunk, options Options, downloaded *atomic.Int64, started time.Time, progress ProgressFunc) error {
	var lastErr error
	for attempt := 0; attempt < options.Retries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(time.Duration(attempt) * 300 * time.Millisecond):
			}
		}
		err := fetchChunk(ctx, client, file, source, part, options, downloaded, started, progress)
		if err == nil {
			return nil
		}
		lastErr = err
	}
	return fmt.Errorf("%s bytes %d-%d failed: %w", source.Name, part.start, part.end, lastErr)
}
